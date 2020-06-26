GIT_VERSION?=$(shell git describe --tags --always --abbrev=42 --dirty)

build: bin
	go build -o bin/sidecar  \
		-ldflags "-X github.com/factorysh/traefik-sidecar/version.version=$(GIT_VERSION)" \
		.

bin:
	mkdir -p bin

docker-build:
	mkdir -p .cache/go-pkg
	docker run --rm -ti \
		-v `pwd`:/src \
		-w /src \
		-v `pwd`/.cache:/.cache \
		-v `pwd`/.cache/go-pkg:/go/pkg \
		-u `id -u` \
		bearstech/golang-dev \
		make build

docker-upx:
	docker run -ti --rm \
		-u `id -u` \
		-v `pwd`/bin:/upx \
		-w /upx \
		bearstech/upx \
		upx sidecar
