# Compile the Proto3 File into Golang
protoc --go_out=. --go_opt=paths=source_relative proto/messages.proto

# Move the compiled code into the backend source
mv ./proto/messages.pb.go ./backend/pkg/api/messages.pb.go

# Compile the protobuf to javascript
npx pbjs ./proto/messages.proto --es6 ./frontend/src/network/message.pb.js


