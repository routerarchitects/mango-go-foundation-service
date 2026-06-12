# Mango Cloud Go Foundation Service

A standardized, production-ready microservice foundation skeleton for the Mango Cloud (OpenWiFi) environment. This service provides a pre-configured architecture featuring dual-port HTTP server separation, PostgreSQL integration, automated schema migrations, and Service Discovery out-of-the-box.

---

## Folder Structure

```text
├── .github/
│   └── workflows/
│       └── ci.yaml              # Continuous Integration workflow configuration
├── cmd/
│   └── main.go                  # Service startup coordinator and lifecycle manager
├── db/
│   └── schema/                  # SQL schema migrations directory
│       └── 0001_initial.sql     # Placeholder SQL table setup
├── docs/                        # Specifications and API contracts templates
│   ├── requirements.md          # Requirements template
│   ├── design.md                # Technical design doc template
│   └── openapi.yaml             # OpenAPI (Swagger) api definition
├── configs/                     # Configurations for development/testing
│   └── local-dev.env            # Env configuration for local running (outside Docker)
├── deployments/                 # Deployment-related configurations
│   └── docker-compose/
│       ├── docker-compose.env   # Env template for Docker Compose execution
│       └── docker-compose.yaml  # Docker Compose deployment integration template
├── external/                    # Third-party API client integration wrappers
│   └── README.md                # Developer guide for external adapters
├── internal/
│   ├── config/                  # caarlos0/env environment parsing
│   ├── db/                      # Connection pool (pgxpool) & migration engine
│   ├── http/                    # Routing, middleware, and Dual TLS engine
│   ├── models/                  # Domain-level request/response model structs
│   └── services/                # Business logic interfaces and services
├── .dockerignore                # Exclusions for Docker build context
├── .gitignore                   # Exclusions for Git repository
├── Dockerfile                   # Multi-stage production container configuration
├── init-service.sh              # Scaffolding helper script to rename/configure
├── Makefile                     # Build, run, test, and containerize commands
└── README.md                    # This developer guide
```

---

## Scaffolding a New Service

To instantiate a new service using this foundation template:

1. Execute the `init-service.sh` script, providing your new service name, public API port, private API port, and target directory:
   ```bash
   ./init-service.sh <new-service-name> <public-port> <private-port> [target-directory]
   ```

2. **Example**:
   ```bash
   ./init-service.sh mango-go-foundation-service 16012 17012 ../mango-go-foundation-service
   ```

3. Navigate to the generated directory and start customizing.

---

## Local Development (Outside Docker)

### Prerequisites
* Go 1.25+ installed
* PostgreSQL and Kafka running (or forwarded to `localhost`)

### Steps
1. Populate certificates under a `./certs` directory in your workspace:
   * `./certs/restapi-cert.pem`
   * `./certs/restapi-key.pem`
   * `./certs/restapi-ca.pem`
   *(Self-signed certificates generated during container launch can be copied over).*

2. Source the local dev environment variables and run:
   ```bash
   make run
   # OR: source configs/local-dev.env && go run ./cmd
   ```

---

## Docker Integration

### 1. Build the Image
```bash
make docker-build
```

### 2. Integrate with Mango Cloud Compose Stack
1. Copy the generated service env file to the `mango-cloud-deployment/docker-compose/` folder:
   ```bash
   cp deployments/docker-compose/<your-service-name>.env /openwifi-sdk/mango-cloud-deployment/docker-compose/
   ```

2. Copy the pre-configured service block from `deployments/docker-compose/docker-compose.yaml` and paste it inside the `services:` declaration of `/openwifi-sdk/mango-cloud-deployment/docker-compose/docker-compose.yml`.

3. Re-launch the compose deployment:
   ```bash
   docker compose up -d --build <your-service-name>
   ```

---

## Database Migrations
Migrations are managed dynamically. When the service boots up:
1. It validates the database connection.
2. It verifies the presence of the `schema_migrations` tracking table.
3. It scans the `db/schema/` directory for `.sql` files.
4. Any SQL files that have not been registered are executed sequentially in individual SQL transactions.
5. If a migration fails, the transaction is rolled back and the service blocks startup to prevent running on a broken schema.
