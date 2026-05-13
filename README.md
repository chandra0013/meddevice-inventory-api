# meddevice-inventory-api

RESTful medical device inventory API built with Go, PostgreSQL, Docker, and GitHub Actions CI.

## Why this project exists

Medical device teams need a reliable way to track device status, location, and lifecycle state across clinical environments. This API provides a small production-style backend for device inventory operations with clear REST endpoints, API-key authentication, PostgreSQL persistence, containerized local development, and CI coverage.

## Architecture

```text
Client
  |
  | Authorization: Bearer <API_KEY>
  v
Go HTTP API
  |
  v
PostgreSQL
```

## Tech stack

- Go 1.22
- PostgreSQL
- Docker and Docker Compose
- GitHub Actions CI
- Standard library HTTP router

## Endpoints

| Method | Path | Description |
| --- | --- | --- |
| GET | `/health` | Health check |
| GET | `/devices?limit=20&offset=0` | List devices with pagination |
| POST | `/devices` | Create a device |
| GET | `/devices/{id}` | Get one device |
| PUT | `/devices/{id}` | Update one device |
| DELETE | `/devices/{id}` | Delete one device |

All `/devices` endpoints require either `Authorization: Bearer <API_KEY>` or `X-API-Key: <API_KEY>`.

## Device payload

```json
{
  "name": "Cardiac Monitor",
  "manufacturer": "Acme MedTech",
  "serial_number": "CM-100",
  "status": "in_service",
  "location": "ICU"
}
```

`status` must be one of `in_service`, `maintenance`, or `retired`.

## Run locally

```bash
docker compose up --build
```

The API will be available at `http://localhost:8080`.

## Example request

```bash
curl -X POST http://localhost:8080/devices \
  -H "Authorization: Bearer dev-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Infusion Pump",
    "manufacturer": "Acme MedTech",
    "serial_number": "IP-2048",
    "status": "in_service",
    "location": "Ward A"
  }'
```

## Environment variables

| Name | Required | Default |
| --- | --- | --- |
| `ADDR` | No | `:8080` |
| `API_KEY` | Yes | `dev-api-key` for local development |
| `DATABASE_URL` | Yes | Local PostgreSQL compose URL |

## Development

```bash
go mod download
go test ./...
go build ./cmd/api
```

## Repository description

Use this GitHub description:

```text
RESTful medical device inventory API - Go, PostgreSQL, Docker, CI/CD pipeline
```
