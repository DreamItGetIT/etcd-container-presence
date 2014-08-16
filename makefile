.PHONY: default docker-image build-register

deafult: build-register docker-image

docker-image:
	@echo "Building docker image"
	@docker build -t digit/etcd-container-presence .

build-register:
	@echo "Compiling register"
	@gom build -o docker-image/register
