all:
	go run main.go

build: clean
	mkdir -p ./build
	GOOS=linux GOARCH=amd64 go build -o build/bootstrap main.go
	zip ./build/bootstrap.zip ./build/bootstrap

deploy:
	cd infrastructure/ && tofu init && tofu apply

clean: 
	rm -rf build
