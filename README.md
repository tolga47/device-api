# device-api

A small REST service for managing devices, built with Go, Gin and SQLite.

## Running

```bash
go run .
```

The server listens on `:8080`. Data is stored in `devices.db` by default; set `DB_PATH` to use a different file.

```bash
DB_PATH=/tmp/devices.db go run .
```

## Running with Docker

```bash
docker build -t device-api .
docker run -p 8080:8080 -v device-data:/data device-api
```

The database lives in the `device-data` volume, so it survives container restarts.

## Tests

```bash
go test ./...
```

## Endpoints

| Method | Path                   | Description                        |
|--------|------------------------|------------------------------------|
| POST   | `/devices`             | Create a device                    |
| GET    | `/devices`             | List all devices                   |
| GET    | `/devices?brand=apple` | Search devices by brand            |
| GET    | `/devices/{id}`        | Get a device                       |
| PUT    | `/devices/{id}`        | Replace a device (name and brand)  |
| PATCH  | `/devices/{id}`        | Update only the fields you send    |
| DELETE | `/devices/{id}`        | Delete a device                    |

Brand search is case-insensitive. `id` and `created_at` are set by the server and can't be changed.

Errors come back as `{"error": "..."}` with `400` for bad input and `404` when the device doesn't exist.

## Examples

```bash
# create
curl -X POST localhost:8080/devices \
  -H "Content-Type: application/json" \
  -d '{"name": "iPhone 15", "brand": "Apple"}'

# response: 201 Created, Location: /devices/1
# {"id":1,"name":"iPhone 15","brand":"Apple","created_at":"2026-11-08T21:53:27.381792Z"}

# get
curl localhost:8080/devices/1

# search by brand
curl "localhost:8080/devices?brand=apple"

# full update
curl -X PUT localhost:8080/devices/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "iPhone 16", "brand": "Apple"}'

# partial update
curl -X PATCH localhost:8080/devices/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "iPhone 16 Pro"}'

# delete
curl -X DELETE localhost:8080/devices/1
```
