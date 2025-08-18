package db

import (
	"archive/zip"
	"context"
	"elevation"
	"sync"
)

type freq struct {
	mu sync.RWMutex
	m  map[string]int
}

func newFreq() *freq {
	return &freq{
		m: make(map[string]int),
	}
}

func (f *freq) Get(name string) (int, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	i, ok := f.m[name]
	return i, ok
}

// if exists add one, otherwise set to 1
func (f *freq) AddOne(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.m[name]; ok {
		f.m[name]++
		return
	}
	f.m[name] = 1
}

func (f *freq) Least() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var least string
	leastCnt := -1
	for name, cnt := range f.m {
		if leastCnt == -1 {
			least = name
			leastCnt = cnt
			continue
		}
		if cnt < leastCnt {
			least = name
			leastCnt = cnt
		}
	}
	return least
}

const (
	MB = 1 << 20
	GB = 1 << 30
	TB = 1 << 40
)

type FsMemCacheDb struct {
	mu sync.RWMutex
	// path to data
	root string
	// how much should be allowed to be kept in the cache (in bytes)
	maxMemorySize int64
	// based on maxMemorySize; each files is 25mb zipped
	maxNumInCache int
	cache         map[string][]int16
	freq          *freq
}

func newFsMemCacheDb(root string, maxMemSize int64) *FsMemCacheDb {
	return &FsMemCacheDb{
		root:          root,
		maxMemorySize: maxMemSize,
		maxNumInCache: int(maxMemSize / 25 * MB),
		cache:         make(map[string][]int16),
		freq:          newFreq(),
	}
}

func (db *FsMemCacheDb) ReadElevation(ctx context.Context, resolution elevation.Resolution, lat float64, lng float64) (elevation.HGTRecord, error) {
	name := elevation.LatLngToTileName(lat, lng)
	tileLat, tileLng, err := elevation.ParseTileName(name)
	if err != nil {
		return elevation.HGTRecord{}, err
	}

	// the simple case is that the data is already in the cache
	if data, exists := db.cache[name]; exists {
		db.mu.RLock()
		elev, err := elevation.GetElevation(data, resolution, tileLat, tileLng, lat, lng)
		if err != nil {
			return elevation.HGTRecord{}, err
		}
		db.mu.RUnlock()
		db.freq.AddOne(name)
		return elevation.HGTRecord{
			Latitude:  lat,
			Longitude: lng,
			Elevation: elev,
		}, nil
	}

	// we know data is not in the cache
	zr, err := zip.OpenReader(db.root + "/" + name + ext)
	if err != nil {
		return elevation.HGTRecord{}, err
	}
	data, tileLat, tileLng, err := elevation.GetDataFromZip(zr, resolution)
	if err != nil {
		return elevation.HGTRecord{}, err
	}
	elev, err := elevation.GetElevation(data, resolution, tileLat, tileLng, lat, lng)
	if err != nil {
		return elevation.HGTRecord{}, err
	}

	// not in cache and cache is not full
	if len(db.cache) < db.maxNumInCache {
		db.mu.Lock()
		db.cache[name] = data
		db.mu.Unlock()
		db.freq.AddOne(name)
		return elevation.HGTRecord{Latitude: lat, Longitude: lng, Elevation: elev}, nil
	}

	// not in cache and cache is full
	delete(db.cache, db.freq.Least())
	db.mu.Lock()
	db.cache[name] = data
	db.mu.Unlock()
	db.freq.AddOne(name)
	return elevation.HGTRecord{}, nil
}
