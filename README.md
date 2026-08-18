# Mango Cloud Go Foundation Service

A standardized, production-ready microservice foundation skeleton for the Mango Cloud (OpenWiFi) environment. This service provides a pre-configured architecture featuring dual-port HTTP server separation, PostgreSQL integration, automated schema migrations, and Service Discovery out-of-the-box.

## About Mango Cloud

[Mango Cloud](https://www.mangowifi.cloud/) is Router Architects' open-source platform for managed Wi-Fi and connectivity operations. Purpose built for ISPs, MSPs, Edge AI and IoT service provider.

Built around the OpenLAN and OpenWiFi ecosystem, Mango Cloud provides a cloud operations layer for provisioning, monitoring, device lifecycle management, subscriber and tenant workflows, analytics, topology, and provider integrations across access points, gateways, switches, and CPE.

## Role of this Repository

`mango-go-foundation-service` provides the standardized Go microservice foundation used to build new Mango Cloud backend services.

It provides reusable patterns for:

- HTTP service architecture
- PostgreSQL integration
- Database migrations
- Service discovery
- Configuration management
- Docker deployment
- API specification
- CI workflows
- Testing and technical documentation

New Mango Cloud services can use this repository as a common starting point so that services follow consistent architecture, deployment, and operational conventions.

- [Mango Cloud](https://www.mangowifi.cloud/)
- [Mango Cloud Deployment](https://github.com/routerarchitects/mango-cloud-deployment)
- [Router Architects](https://www.routerarchitects.com/)
  
---

## Folder Structure

```text
├── .github/
│   └── workflows/
│       └── ci.yaml              # Continuous Integration workflow configuration
├── cmd/
│   └── main.go                  # Boilerplate entrypoint (Config load, Logger init, runs App, OS signals)
├── db/
│   └── schema/                  # SQL schema migrations directory
│       └── 0001_initial.sql     # Placeholder SQL table setup
├── docs/                        # Specifications and API contracts templates
│   ├── phase-1/                 # Phase 1 documentation deliverables
│   │   ├── design.md            # Technical design doc template
│   │   ├── openapi.yaml         # OpenAPI (Swagger) API definition
│   │   └── testcases.md         # API test cases and validation matrices
│   └── requirements.md          # Functional & non-functional requirements template
├── configs/                     # Configurations for development/testing
│   └── local-dev.env            # Env configuration for local running (outside Docker)
├── deployments/                 # Deployment-related configurations
│   └── docker-compose/
│       ├── docker-compose.env   # Env template for Docker Compose execution
│       └── docker-compose.yaml  # Docker Compose deployment integration template
├── external/                    # Third-party API client integration wrappers
│   └── README.md                # Developer guide for external adapters
├── internal/
│   ├── app/                     # Application wiring and dependency injection
│   │   └── app.go               # Dynamic struct creation, DB pool, and module boot
│   ├── config/                  # caarlos0/env environment parsing
│   ├── db/                      # Connection pool (pgxpool) & migration engine
│   ├── http/                    # Routing, middleware, and Dual TLS engine
│   ├── models/                  # Domain-level request/response model structs
│   └── services/                # Business logic interfaces and services
├── .dockerignore                # Exclusions for Docker build context
├── .gitignore                   # Exclusions for Git repository
├── Dockerfile                   # Multi-stage production container configuration
├── Makefile                     # Build, run, test, and containerize commands
├── README.md                    # This developer guide
```

---

## Phase 1: Scaffolding a New Service

To initialize a new repository using this foundation template:

1. **Clone both the template and your new repository in a workspace:**
   ```bash
   mkdir <workspace-dir>
   cd <workspace-dir>
   git clone git@github.com:routerarchitects/mango-go-foundation-service.git
   git clone git@github.com:routerarchitects/<new-service-name>.git
   ```

2. **Copy the template files into your new repository:**
   ```bash
   cd <new-service-name>
   git checkout -b base-service-scaffold
   cp -rf ../mango-go-foundation-service/!(.git|.idea|.vscode|tmp|bin) .
   ```

3. **Commit and push the scaffold as the first commit:**
   ```bash
   git add .
   git commit -m "Initial service scaffold"
   git push origin base-service-scaffold
   ```

4. **Merge the base scaffold to your main branch**:
   To establish a clean baseline in your repository:
   * Open a Pull Request (PR) on GitHub from `base-service-scaffold` to your main branch (e.g., `main` or `master`).
   * Review and merge the PR.
   * Switch back to your local main branch and pull the merged changes:
     ```bash
     git checkout main
     git pull origin main
     ```

---

## Phase 2: Configuring your New Service

Once the clean base scaffold is merged into your main branch (Phase 1), create a new configuration branch to customize the template:

1. **Create a customization branch**:
   ```bash
   git checkout -b configure-service
   ```

2. **Customize the service name and ports**:
   Define your service settings as environment variables, then run the customization and rename commands:
   ```bash
   # 1. Define your new service configurations (e.g. PUBLIC_PORT="16010", PRIVATE_PORT="17010"):
   export NEW_SERVICE_NAME="<new-service-name>"
   export PUBLIC_PORT="<public-port>"
   export PRIVATE_PORT="<private-port>"

   # 2. Customize all files using the variables:
   find . -type f -not -path '*/.git/*' -exec sed -i \
       -e "s/{{SERVICE_NAME}}/${NEW_SERVICE_NAME}/g" \
       -e "s/mango-go-foundation-service/${NEW_SERVICE_NAME}/g" \
       -e "s/{{PUBLIC_PORT}}/${PUBLIC_PORT}/g" \
       -e "s/{{PRIVATE_PORT}}/${PRIVATE_PORT}/g" {} +

   # 3. Rename the compose environment file:
   mv deployments/docker-compose/docker-compose.env deployments/docker-compose/${NEW_SERVICE_NAME}.env
   ```

3. **Commit and push your customization changes**:
   ```bash
   git add .
   git commit -m "refactor: rename service and customize ports"
   git push origin configure-service
   ```
   * Open a Pull Request from `configure-service` to your main branch.
   * Merge the PR to complete the service initialization!

---

## Phase 3: Docker Integration

### 1. Build the Image
```bash
make docker-build
```

### 2. Integrate with Mango Cloud Compose Stack
1. Copy the customized environment file manually:
   ```bash
   cp deployments/docker-compose/${NEW_SERVICE_NAME}.env /path_to/mango-cloud-deployment/docker-compose/
   ```

2. Paste the service block from `deployments/docker-compose/docker-compose.yaml` inside the `services:` block of your deployment's `docker-compose.yml`.

3. Re-launch the compose deployment:
   ```bash
   docker compose up
   ```


