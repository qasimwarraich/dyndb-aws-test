all:
	go run main.go

build: clean
	mkdir -p ./build
	GOOS=linux GOARCH=amd64 go build -o build/bootstrap main.go
	zip ./build/bootstrap.zip ./build/bootstrap

deploy: build
	cd infrastructure/ && tofu init && tofu apply

destroy: 
	cd infrastructure/ && tofu init && tofu destroy

clean: 
	rm -rf build
