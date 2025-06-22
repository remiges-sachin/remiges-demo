#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Remiges Demo Cleanup Script ===${NC}\n"

# Function to confirm action
confirm() {
    local prompt=$1
    local default=${2:-N}
    
    if [ "$default" = "Y" ]; then
        prompt="$prompt [Y/n]: "
    else
        prompt="$prompt [y/N]: "
    fi
    
    read -p "$prompt" response
    response=${response:-$default}
    
    case "$response" in
        [yY][eE][sS]|[yY]) 
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

# Script configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Change to project root
cd "$PROJECT_ROOT"

# Step 1: Stop running application
echo -e "${YELLOW}Step 1: Stopping any running application processes${NC}"
pkill -f "go run" 2>/dev/null
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Application processes stopped${NC}"
else
    echo -e "${GREEN}✓ No application processes found${NC}"
fi

# Step 2: Stop Docker services
echo -e "\n${YELLOW}Step 2: Stopping Docker services${NC}"
cd "$PROJECT_ROOT"
docker-compose down
echo -e "${GREEN}✓ Services stopped${NC}"

# Step 3: Clean up volumes (optional)
if confirm "\nDo you want to remove Docker volumes (this will delete all data)?"; then
    echo -e "${YELLOW}Removing Docker volumes...${NC}"
    docker-compose down -v
    echo -e "${GREEN}✓ Volumes removed${NC}"
fi

# Step 4: Clean up generated files (optional)
if confirm "\nDo you want to remove generated files (logs, etc.)?"; then
    echo -e "${YELLOW}Removing generated files...${NC}"
    rm -f app.log session-*.log 2>/dev/null
    echo -e "${GREEN}✓ Generated files removed${NC}"
fi

# Step 5: Clean up Go build cache (optional)
if confirm "\nDo you want to clean Go build cache?"; then
    echo -e "${YELLOW}Cleaning Go build cache...${NC}"
    go clean -cache
    echo -e "${GREEN}✓ Go build cache cleaned${NC}"
fi

# Summary
echo -e "\n${GREEN}=== Cleanup Complete ===${NC}"
echo -e "\nTo reinitialize the project, run:"
echo "  ./scripts/init.sh"