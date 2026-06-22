# Common Liveness API Test Cases

Scope: Common liveness endpoints defined in `openapi.yaml`.

This document outlines the test case tables for the liveness endpoint shared across Mango Cloud microservices.

## Global Test Assumptions

- The service exposes a dual-port HTTP architecture.
- The `/livez` route is accessible, unauthenticated, on both ports.

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
