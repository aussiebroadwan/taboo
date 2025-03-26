#!/bin/bash 

# Check if swag is installed
if ! command -v swag >/dev/null 2>&1; then
    echo "Error: swag is not installed. Please install the swag application. See more https://github.com/swaggo/swag"
    exit 1
fi

# Changing to the backend directory to where the go.mod file is
cd backend

# Compling swagger docs to backend/docs/ 
swag init -o ./docs -g ./pkg/web/swagger.go

# Restoring the current directory
cd ..
