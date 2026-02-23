# Taboo development commands

# Build frontend assets
build-frontend:
    cd frontend && npm install && npm run build
    find internal/frontend/dist -type f ! -name '.gitkeep' -delete 2>/dev/null || true
    find internal/frontend/dist -mindepth 1 -type d -delete 2>/dev/null || true
    cp -r frontend/dist/* internal/frontend/dist/

# Build everything (frontend + backend)
build: build-frontend
    go build -o bin/taboo ./cmd/taboo

# Build with version info (includes frontend)
build-release: build-frontend
    go build -ldflags "-X github.com/aussiebroadwan/taboo/internal/app.Version=$(git describe --tags --always --dirty) -X github.com/aussiebroadwan/taboo/internal/app.Commit=$(git rev-parse HEAD) -X github.com/aussiebroadwan/taboo/internal/app.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o bin/taboo ./cmd/taboo

# Run tests
test:
    go test -v -race ./...

# Run linter
lint:
    golangci-lint run
    cd frontend && npm run lint

# Generate SQLC code
generate:
    cd internal/store/drivers/sqlite && sqlc generate

# Format code
fmt:
    go fmt ./...

# Run with debug logging
dev:
    go run ./cmd/taboo serve --log-level debug

# Run with config file
run: build
    ./bin/taboo serve -c config.yaml

# Clean build artifacts
clean:
    rm -rf bin/ coverage.out coverage.html taboo.db
    find internal/frontend/dist -type f ! -name '.gitkeep' -delete 2>/dev/null || true
    find internal/frontend/dist -mindepth 1 -type d -delete 2>/dev/null || true
