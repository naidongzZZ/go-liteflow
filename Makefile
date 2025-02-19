
proto: 
	protoc --go_out=pb --go-grpc_out=pb pb/core.proto

co:
	go run cmd/main.go --mode co --addr 127.0.0.1:20021

tm:
	go run cmd/main.go --mode tm --addr 127.0.0.1:20022 --co_addr 127.0.0.1:20021

clean:
	rm -rf /tmp/task_ef
