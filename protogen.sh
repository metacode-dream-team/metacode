# WARNING CHANGE PATH ACCORDING TO WHAT PROTO YOU WANT TO GENERATE
protoc -I proto \
    --go_out=pkg/pb --go_opt=paths=source_relative \
    --go-grpc_out=pkg/pb --go-grpc_opt=paths=source_relative \
    notification/notification.proto