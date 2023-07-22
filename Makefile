PUBLIC_REGISTRY_HOST=docker.io
PUBLIC_REGISTRY_OWNER=bulvinkl
PUBLIC_REGISTRY_APP_NAME=cbr-currency-converter

CI_COMMIT_REF_NAME=latest

all: deps build test

deps:
	@go mod download
	@echo "Dependencies installed successfully"

build:
	mkdir -p build
	go build -o build/cbr cmd/*.go

clean:
	rm -rf build

test:
	go test -v -covermode=count './...'

image:
	docker build -t ${PUBLIC_REGISTRY_HOST}/${PUBLIC_REGISTRY_OWNER}/${PUBLIC_REGISTRY_APP_NAME}:${CI_COMMIT_REF_NAME} ./
	docker push ${PUBLIC_REGISTRY_HOST}/${PUBLIC_REGISTRY_OWNER}/${PUBLIC_REGISTRY_APP_NAME}:${CI_COMMIT_REF_NAME}
