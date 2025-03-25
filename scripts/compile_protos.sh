#!/bin/bash

# Check if protoc is installed
if ! command -v protoc >/dev/null 2>&1; then
    echo "Error: protoc is not installed. Please install the Protocol Buffers compilera. See more https://grpc.io/docs/protoc-installation/"
    exit 1
fi

# Check if npx is installed
if ! command -v npx >/dev/null 2>&1; then
    echo "Error: npx is not installed. Please install Node.js and npm. See more https://nodejs.org/en/download"
    exit 1
fi

# Compile the Proto3 File into Golang
protoc --go_out=. --go_opt=paths=source_relative proto/messages.proto

# Move the compiled code into the backend source
mv ./proto/messages.pb.go ./backend/pkg/web/messages.pb.go

# Compile the protobuf to javascript
npx pbjs ./proto/messages.proto --es6 ./frontend/src/network/message.pb.js


