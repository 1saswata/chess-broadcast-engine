gen:
	protoc --go_out=. --go_opt=module=github.com/1saswata/chess-broadcast-engine --go-grpc_out=.  --go-grpc_opt=module=github.com/1saswata/chess-broadcast-engine proto/chess_match.proto
test: 
	go test -v ./...
