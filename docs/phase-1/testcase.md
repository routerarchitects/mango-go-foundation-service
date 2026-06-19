# Common System & Diagnostics API Test Cases

Scope: Common system, diagnostics, and liveness endpoints defined in `openapi.yaml`.

This document outlines the test case tables for standard endpoints shared across Mango Cloud microservices.

## Global Test Assumptions

- The service exposes a dual-port HTTP architecture:
  1. **Public Port**: Exposes business-level API endpoints (requires bearer/JWT token validation).
  2. **Private Port**: Exposes system management and diagnostics (requires `X-INTERNAL-NAME` and `X-API-KEY` headers).
- The `/livez` route is accessible, unauthenticated, on both ports.
- The `/api/v1/system` routes (GET and POST) are private admin endpoints and are restricted to the private port only.

---

# API: Liveness Probe

```http
GET /livez
```

Purpose: Verify the microservice runtime is active and healthy.

Security: Unauthenticated (accessible on both public and private ports).

## Test Cases

| ID | Name | Expected Result |
|---|---|---|
| TC-LIVEZ-001 | Health check returns healthy | `200 OK`; indicating runtime is fully operational |
| TC-LIVEZ-002 | Health check returns unhealthy | `500 Internal Server Error`; when internal dependencies or critical threads are unresponsive |
| TC-LIVEZ-003 | Access health check unauthenticated | `200 OK` (or `500`); request succeeds without requiring `X-API-KEY` or Bearer tokens |

---

# API: Retrieve System Diagnostics

```http
GET /api/v1/system
```

Purpose: Fetch system resource usage or static status diagnostics info.

Security: Requires `PrivateInternalNameAuth` (`X-INTERNAL-NAME`) and `PrivateApiKeyAuth` (`X-API-KEY`). Restricted to the private port.

Required query parameters:
* `command` (Enum: `info`, `resources`)

## Test Cases

| ID | Name | Expected Result |
|---|---|---|
| TC-SYS-GET-001 | Retrieve static system info successfully | `200 OK`; returns system version, build information, and environment scope |
| TC-SYS-GET-002 | Retrieve resource statistics successfully | `200 OK`; returns active CPU/Memory usage metrics and DB pool statistics |
| TC-SYS-GET-003 | Missing required `command` query parameter | `400 Bad Request`; returns validation/bad request error |
| TC-SYS-GET-004 | Invalid `command` query parameter value | `400 Bad Request`; returns validation/bad request error |
| TC-SYS-GET-005 | Missing authentication headers (`X-INTERNAL-NAME` or `X-API-KEY`) | `401 Unauthorized`; returns unauthorized error |
| TC-SYS-GET-006 | Invalid authentication headers | `401 Unauthorized`; returns unauthorized error |
| TC-SYS-GET-007 | Access GET system API via public port | `404 Not Found` or `403 Forbidden` (private admin route is unreachable) |

---

# API: Modify Log Levels or Query Diagnostics Schema

```http
POST /api/v1/system
```

Purpose: Dynamically change subsystem logging levels or inspect logging schema/metadata.

Security: Requires `PrivateInternalNameAuth` (`X-INTERNAL-NAME`) and `PrivateApiKeyAuth` (`X-API-KEY`). Restricted to the private port.

Request body example (Set log level):
```json
{
  "command": "setloglevel",
  "subsystems": [
    {
      "tag": "http",
      "value": "debug"
    }
  ]
}
```

Request body example (Get subsystem names):
```json
{
  "command": "getsubsystemnames"
}
```

## Test Cases

| ID | Name | Expected Result |
|---|---|---|
| TC-SYS-POST-001 | Set log level successfully for a subsystem | `200 OK`; log level updated at runtime; logs reflect change |
| TC-SYS-POST-002 | Get current active log levels | `200 OK`; returns mapping of subsystems to their active log levels |
| TC-SYS-POST-003 | Get allowed log level names | `200 OK`; returns list of valid level strings (e.g. debug, info, warn, error) |
| TC-SYS-POST-004 | Get allowed subsystem tags | `200 OK`; returns list of all configurable subsystem tags |
| TC-SYS-POST-005 | Missing required `command` field in request body | `400 Bad Request`; returns validation error |
| TC-SYS-POST-006 | Invalid `command` value | `400 Bad Request`; returns validation error |
| TC-SYS-POST-007 | Missing `subsystems` array for `setloglevel` command | `400 Bad Request`; returns validation error |
| TC-SYS-POST-008 | Unknown fields in request JSON body | `400 Bad Request`; returns validation error |
| TC-SYS-POST-009 | Missing or invalid authentication headers | `401 Unauthorized`; returns unauthorized error |
| TC-SYS-POST-010 | Access POST system API via public port | `404 Not Found` or `403 Forbidden` |
