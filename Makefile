include .env

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=go-insight
OUT_DIR=bin

PROD=prod
DEV=dev
LDFLAGS_SDK=-X 'main.BaseUrl=$(BaseUrl)' -X 'main.PublicKey=$(PublicKey)' -X 'main.PrivateKey=$(PrivateKey)' -X 'main.OrgId=$(OrgId)'

LDFLAGS="-X 'main.ENV=$(PROD)' $(LDFLAGS_SDK)"
LDFLAGS_DEV="-X 'main.ENV=$(DEV)' $(LDFLAGS_SDK)"

all: build

build:
	mkdir -p bin
	$(GOBUILD) -o $(OUT_DIR)/$(BINARY_NAME) -v -ldflags $(LDFLAGS)

build_dev:
	$(GOBUILD) -o $(OUT_DIR)/$(BINARY_NAME) -v -ldflags $(LDFLAGS_DEV)

clean:
	rm -rf ./$(OUT_DIR)

run: build
	./$(OUT_DIR)/$(BINARY_NAME)