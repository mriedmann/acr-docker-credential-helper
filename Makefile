BINARY := docker-credential-acr
IMAGE  := docker-credential-acr:dev

.PHONY: build install image test clean

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY) .

install: build
	cp $(BINARY) ~/bin/ && chmod +x ~/bin/$(BINARY)

image:
	podman build -t $(IMAGE) .

test:
	go vet ./... && go test -v -race ./...

clean:
	rm -f $(BINARY)
