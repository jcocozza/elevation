# Elevation

A golang package for working with srtm data.

## Data Download

A list of datafiles can be found in thie [srtm30m_urls.txt](data/srtm30m_urls.txt).

All the data can be downloaded with the [get.sh](data/get.sh).

You will need to have a nasa earth data account.
See [their site](https://urs.earthdata.nasa.gov/documentation/for_users/data_access/curl_and_wget) their site for more details.

The tldr, once you create an account, add this to your `.netrc`

```
machine urs.earthdata.nasa.gov
    login <username>
    password <password>
```

Then the `get.sh` script should work.

> [!CAUTION]
> This will download around ~100GB of (zipped) data to your machine.
> Everything unzipped is ~346GB

## Usage

There is a CLI tool or you can use the exported code in your own programs.

The cli can be built with:
```bash
go build ./cmd/elevation
```

## Server

The `pkg/` directory contains code to integrate the data retrieved with the `data/get.sh` script with an http server.

Serving the data:
`elevation serve data/`

### API Routes

A single point can be queried via: `/elevation/point?lat=<latitude>&lng=<longitude>`.
> Currently the only supported interpolation method is bilinear.

Many points can be queried with a post request: to `/elevation/points/`

The post can contain a polyline encoded string, or a list of lat/lng pairs.
If the request includes both, the polyine will be used.

Example polyline:
```sh
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"polyline":"sws~FdgyuO`vAhCfeCiCjhaAcg_@`u`BccnZz`~I{vpb@oh`Tkr`U??qo^_deI{ow@rjtJ"}' \
  http://localhost:8001/elevation/points
```

Example points:
```sh
curl --header "Content-Type: application/json" \
  --request POST \
  --data '@latlng.json' \
  http://localhost:8001/elevation/points
```

Where `latlng.json` looks like this:

```json
{
  "points": [
    [
      41.88554,
      -87.62499
    ],
    [
      43.11119,
      -73.76245
    ]
  ]
}
```

The response for the polyline example:

```json
{
  "elevations": [
    {
      "latitude": 41.88554,
      "longitude": -87.62499,
      "elevation": 217.23606399999196
    },
    {
      "latitude": 41.87161,
      "longitude": -87.62568,
      "elevation": 195.73878399995456
    },
    {
      "latitude": 41.85013,
      "longitude": -87.62499,
      "elevation": 186
    },
    {
      "latitude": 41.51071,
      "longitude": -87.45985,
      "elevation": 190.88799999997764
    },
    {
      "latitude": 41.010540000000006,
      "longitude": -82.95871,
      "elevation": 291.33606399999536
    },
    {
      "latitude": 39.21312,
      "longitude": -77.13345,
      "elevation": 172.58255999994353
    },
    {
      "latitude": 42.6604,
      "longitude": -73.52074999999999,
      "elevation": 369.57999999999265
    },
    {
      "latitude": 42.6604,
      "longitude": -73.52074999999999,
      "elevation": 369.57999999999265
    },
    {
      "latitude": 42.82177,
      "longitude": -71.85082999999999,
      "elevation": 332.4924640000276
    },
    {
      "latitude": 43.11119,
      "longitude": -73.76244999999999,
      "elevation": 94.74800000005075
    }
  ],
  "gain": 302.3335040000445,
  "loss": -424.8215679999857,
  "grade": -0.013869227279388326
}
```
Note that net grade, total gain and total loss are also computed.
