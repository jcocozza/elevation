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

## Usage

There is a CLI tool or you can use the exported code in your own programs.

The cli can be built with:
```bash
go build ./cmd/elevation
```

## SQLite and a Server

The `pkg/` directory contains code to integrate a sqlite database with an http server.

Loading to sqlite:
`elevation load -f sqlite -o elevation.db S15W040.hgt`

Serving the data:
`elevation serve elevation.db`

### API Routes

Currently there are three interpolation modes. The default in bilinear.
**Currently bicubic is broken**
`/elevation/<latitude>/<longitude>?interpolation=nearest`
`/elevation/<latitude>/<longitude>?interpolation=bilinear`
`/elevation/<latitude>/<longitude>?interpolation=bicubic`
