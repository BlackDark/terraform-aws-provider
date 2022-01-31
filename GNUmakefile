HOSTNAME=registry.terraform.io
NAMESPACE=hashicorp
NAME=scaffolding
BINARY=terraform-provider-${NAME}
VERSION=0.3.2
OS_ARCH=darwin_amd64

default: testacc

build:
	go build -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m
