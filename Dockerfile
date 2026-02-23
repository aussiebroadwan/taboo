# =============================================================================
# Stage: webbase
#
# Provides a base environment for building the frontend.
# =============================================================================
FROM node:alpine AS webbase

WORKDIR /usr/src/app

# Copy package.json and package-lock.json for dependency caching
COPY frontend/package.json frontend/package-lock.json ./

# Install dependencies
RUN npm install


# =============================================================================
# Stage: webbuild
#
# Builds the frontend for production.
# =============================================================================
FROM webbase AS webbuild

# Copy source code
COPY ./frontend .

# Build the frontend
RUN npm run build


# =============================================================================
# Stage: backendbase
#
# Build the base image with the necessary tools so that it can be used in the
# build stage without having to install them again.
# =============================================================================
FROM golang:alpine AS backendbase

WORKDIR /usr/src/app

RUN apk update && apk upgrade && apk add --no-cache ca-certificates git \
    && update-ca-certificates

# =============================================================================
# Stage: build
#
# Download project dependencies, build the Go binary with version info.
# =============================================================================
FROM backendbase AS build

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading
# them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the built frontend files into the embedded directory
COPY --from=webbuild /usr/src/app/dist ./internal/frontend/dist/

# Copy all Go source
COPY . .

ARG VERSION=dev

RUN CGO_ENABLED=0 go build -ldflags "\
    -X github.com/aussiebroadwan/taboo/internal/app.Version=${VERSION} \
    -X github.com/aussiebroadwan/taboo/internal/app.Commit=$(git rev-parse HEAD) \
    -X github.com/aussiebroadwan/taboo/internal/app.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o /bin/taboo ./cmd/taboo

# =============================================================================
# Stage: release
#
# Copy the built binary and the necessary certificates to a scratch image to
# reduce the image size.
# =============================================================================
FROM scratch AS release

COPY --from=build /bin/taboo /bin/taboo
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/bin/taboo"]
