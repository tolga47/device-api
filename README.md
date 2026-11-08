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
| GET    | `/health`              | Health check                       |

Brand search is case-insensitive. `id` and `created_at` are set by the server and can't be changed.

Errors come back as `{"error": "..."}` with `400` for bad input and `404` when the device doesn't exist. Using a method a path doesn't support (e.g. `DELETE /devices`) returns `405`.

Leading and trailing spaces in `name` and `brand` are trimmed, so a blank value counts as missing.

## Examples

Run these in order on an empty database. Add `-i` to any command to see the status code and headers.

```bash
# create
curl -X POST localhost:8080/devices \
  -H "Content-Type: application/json" \
  -d '{"name": "iPhone 15", "brand": "Apple"}'
# 201 Created, Location: /devices/1
# {"id":1,"name":"iPhone 15","brand":"Apple","created_at":"2026-11-08T21:53:27.381792Z"}

curl -X POST localhost:8080/devices \
  -H "Content-Type: application/json" \
  -d '{"name": "Galaxy S24", "brand": "Samsung"}'
# 201 Created, Location: /devices/2

# list all
curl localhost:8080/devices
# 200 OK
# [{"id":1,"name":"iPhone 15",...},{"id":2,"name":"Galaxy S24",...}]

# search by brand
curl "localhost:8080/devices?brand=apple"
# 200 OK, only the Apple device
# [{"id":1,"name":"iPhone 15","brand":"Apple",...}]

# get
curl localhost:8080/devices/1
# 200 OK
# {"id":1,"name":"iPhone 15","brand":"Apple",...}

# full update
curl -X PUT localhost:8080/devices/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "iPhone 16", "brand": "Apple"}'
# 200 OK, created_at stays the same
# {"id":1,"name":"iPhone 16","brand":"Apple",...}

# partial update
curl -X PATCH localhost:8080/devices/2 \
  -H "Content-Type: application/json" \
  -d '{"name": "Galaxy S25"}'
# 200 OK, brand is untouched
# {"id":2,"name":"Galaxy S25","brand":"Samsung",...}

# delete
curl -X DELETE localhost:8080/devices/2
# 204 No Content

# health check
curl localhost:8080/health
# 200 OK
# {"status":"ok"}
```

### Error cases

```bash
# device doesn't exist
curl -i localhost:8080/devices/999
# 404 {"error":"device not found"}

# id is not a number
curl -i localhost:8080/devices/abc
# 400 {"error":"invalid id"}

# missing field
curl -i -X POST localhost:8080/devices \
  -H "Content-Type: application/json" \
  -d '{"name": "x"}'
# 400 {"error":"name and brand are required"}

# blank field (spaces only)
curl -i -X POST localhost:8080/devices \
  -H "Content-Type: application/json" \
  -d '{"name": "   ", "brand": "Apple"}'
# 400 {"error":"name and brand are required"}

# broken json
curl -i -X POST localhost:8080/devices \
  -H "Content-Type: application/json" \
  -d 'not json'
# 400 {"error":"invalid json body"}

# PUT needs both fields
curl -i -X PUT localhost:8080/devices/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "x"}'
# 400 {"error":"name and brand are required"}

# PATCH can't set a field to empty
curl -i -X PATCH localhost:8080/devices/1 \
  -H "Content-Type: application/json" \
  -d '{"brand": ""}'
# 400 {"error":"name and brand cannot be empty"}

# method not supported on this path
curl -i -X DELETE localhost:8080/devices
# 405 Method Not Allowed, Allow: GET, POST
```
