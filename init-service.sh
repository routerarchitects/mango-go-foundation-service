#!/bin/bash

# ==============================================================================
# OpenWiFi Go Microservice Scaffolding Initializer
# This script configures the template for a new service name and port bindings.
# ==============================================================================

set -euo pipefail

# Print usage information
usage() {
    echo "Usage: $0 <service-name> <public-port> <private-port> [target-directory]"
    echo ""
    echo "Arguments:"
    echo "  service-name      Alphanumeric lowercase name (e.g. parental-control)"
    echo "  public-port       Public API port (e.g. 16008)"
    echo "  private-port      Private/Internal API port (e.g. 17008)"
    echo "  target-directory  Optional destination folder. If omitted, runs in-place."
    echo ""
    echo "Example:"
    echo "  $0 billing-service 16012 17012 ../billing-service"
    exit 1
}

# Check minimum argument count
if [ "$#" -lt 3 ]; then
    usage
fi

SERVICE_NAME=$(echo "$1" | tr '[:upper:]' '[:lower:]')
PUBLIC_PORT=$2
PRIVATE_PORT=$3
TARGET_DIR=${4:-"."}

# Validate service-name matches alpha-numeric and hyphens only
if [[ ! "$SERVICE_NAME" =~ ^[a-z0-9-]+$ ]]; then
    echo "Error: Service name '$SERVICE_NAME' must only contain lowercase letters, numbers, and hyphens."
    exit 1
fi

# Validate ports are numbers
if [[ ! "$PUBLIC_PORT" =~ ^[0-9]+$ ]] || [[ ! "$PRIVATE_PORT" =~ ^[0-9]+$ ]]; then
    echo "Error: Ports must be positive integers."
    exit 1
fi

# Locate the root of the template (where this script resides)
TEMPLATE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Resolve absolute path for Target Directory
mkdir -p "$TARGET_DIR"
TARGET_DIR="$( cd "$TARGET_DIR" && pwd )"

echo "=== Initializing New Go Service ==="
echo "Service Name: $SERVICE_NAME"
echo "Public Port:  $PUBLIC_PORT"
echo "Private Port: $PRIVATE_PORT"
echo "Target Dir:   $TARGET_DIR"
echo "==================================="

# 1. Copy template files to destination if not running in-place
if [ "$TEMPLATE_DIR" != "$TARGET_DIR" ]; then
    echo "Copying template files to target directory..."
    # Copy files excluding .git, binaries, and temporary files
    rsync -a \
      --exclude='.git' \
      --exclude='bin' \
      --exclude='tmp' \
      --exclude='*.log' \
      --exclude='docker-compose*_data' \
      "$TEMPLATE_DIR/" "$TARGET_DIR/"
fi

# 2. Perform string substitutions across all text files
echo "Applying string substitutions..."
cd "$TARGET_DIR"

# List of files to run substitutions on
FILES_TO_PROCESS=$(find . -type f \( \
    -name "*.go" -o \
    -name "*.mod" -o \
    -name "Makefile" -o \
    -name "Dockerfile" -o \
    -name "*.env" -o \
    -name "*.yaml" -o \
    -name "*.yml" -o \
    -name "*.md" -o \
    -name "*.sql" \
    \) -not -path "*/.git/*" -not -name "init-service.sh")

for file in $FILES_TO_PROCESS; do
    # Replace foundation module path with the new module path
    sed -i "s|github.com/routerarchitects/mango-go-foundation-service|github.com/routerarchitects/$SERVICE_NAME|g" "$file"
    
    # Replace service name placeholder
    sed -i "s|{{SERVICE_NAME}}|$SERVICE_NAME|g" "$file"
    sed -i "s|mango-go-foundation-service|$SERVICE_NAME|g" "$file"
    
    # Replace ports placeholders
    sed -i "s|{{PUBLIC_PORT}}|$PUBLIC_PORT|g" "$file"
    sed -i "s|{{PRIVATE_PORT}}|$PRIVATE_PORT|g" "$file"
done

# 3. Rename environment files to match the service name
echo "Renaming environment files..."
if [ -f "deployments/docker-compose/docker-compose.env" ]; then
    mv "deployments/docker-compose/docker-compose.env" "deployments/docker-compose/$SERVICE_NAME.env"
fi

# 4. Initialize dynamic dependencies (go mod tidy)
echo "Resolving dependencies (go mod tidy)..."
if command -v go &> /dev/null; then
    go mod tidy
else
    echo "Warning: 'go' binary not found. Please run 'go mod tidy' manually in the target folder."
fi

echo ""
echo "Success! Service '$SERVICE_NAME' has been initialized."
echo ""
echo "Next Steps:"
echo "  1. Copy 'deployments/docker-compose/$SERVICE_NAME.env' to your deployment directory:"
echo "     $ cp deployments/docker-compose/$SERVICE_NAME.env /openwifi-sdk/mango-cloud-deployment/docker-compose/"
echo "  2. Append the compose service defined in 'deployments/docker-compose/docker-compose.yaml' to:"
echo "     /openwifi-sdk/mango-cloud-deployment/docker-compose/docker-compose.yml"
echo "  3. Start writing your endpoints in 'internal/http/routes/routes.go' and services in 'internal/services/services.go'."
echo ""
