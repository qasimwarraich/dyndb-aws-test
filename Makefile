all:
	go run main.go

build:
	mkdir -p ./build
	GOOS=linux GOARCH=amd64 go build -o build/bootstrap main.go
	zip ./build/bootstrap.zip ./build/bootstrap

clean: 
	rm -rf build
