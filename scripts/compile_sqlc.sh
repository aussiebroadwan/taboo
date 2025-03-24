#!/bin/bash 


# Check if npx is installed
if ! command -v sqlc >/dev/null 2>&1; then
    echo "Error: sqlc is not installed. Please install sqlc. See more https://docs.sqlc.dev/en/latest/overview/install.html"
    exit 1
fi

# Compile the sqlc migration and queries to golang
sqlc generate
