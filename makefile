.PHONY: all docker-image build-register

all: build-register docker-image

docker-image:
	@echo "Building docker image"
	@cd docker-image; docker build -t docker.internal/etcd-container-presence .

build-register:
	@echo "Compiling register"
	@gom build -o docker-image/register
