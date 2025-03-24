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

RUN apk update && apk upgrade && apk add --no-cache ca-certificates \
    && update-ca-certificates

# =============================================================================
# Stage: backendbuild
#
# Download project dependencies, generate sqlc and swag, and build the project.
# =============================================================================
FROM backendbase AS build

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading 
# them in subsequent builds if they change
COPY backend/go.mod backend/go.sum ./
RUN go mod download && go mod verify

# Copy the built frontend files from the frontend stage
COPY --from=webbuild /usr/src/app/dist ./dist

# copy the rest of the source code and build
COPY ./backend .

RUN CGO_ENABLED=0 go build -v -o /bin/tabo .


# =============================================================================
# Stage: release
#
# Copy the built binary and the necessary certificates to a scratch image to
# reduce the image size.
# =============================================================================
FROM scratch AS release

COPY --from=build /bin/tabo /bin/tabo
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/bin/tabo"]

