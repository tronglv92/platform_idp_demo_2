include .env
OS := $(shell uname)
GOPATH:=$(shell go env GOPATH)
TARGET := $(firstword $(MAKECMDGOALS))
OPTIONAL_PARAM := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
API_PROTO_FILES=$(shell find api -name *.proto)
EXT_PROTO_FILES=$(shell find external -name *.proto)

# Export environment variables based on OS
ifeq ($(OS),Darwin)
    export $(shell sed 's/=.*//' .env)
else
		export
endif


.PHONY: init
# init env
init:
	go install github.com/zeromicro/go-zero/tools/goctl@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	go install github.com/favadi/protoc-go-inject-tag@latest
	go install dev.azure.com/PHARMACITYJSC/PharmacityLibrary/_git/pmc-inapp-helper.git/cmd/protoc-gen-go-http@latest

# Proto include paths
PROTO_PATHS := --proto_path=. --proto_path=./helper/protos --proto_path=./helper

# Add local plugins to PATH
export PATH := $(shell pwd)/helper/cmd/protoc-gen-go-http:$(PATH)

.PHONY: add
# add new proto file
add:
	goctl rpc -o api/$(path)/$(path).proto

.PHONY: grpc
# generate grpc code
grpc:
	protoc 	$(PROTO_PATHS) \
			--go_out=paths=source_relative:. \
			--go-grpc_out=paths=source_relative:. \
			--go-http_out=paths=source_relative:. \
	       	--openapi_out=fq_schema_naming=true,default_response=false:. \
			$(API_PROTO_FILES)
	protoc-go-inject-tag -input="./api/*/*.pb.go"

.PHONY: validate
# generate validate code
validate:
	protoc $(PROTO_PATHS) \
           --go_out=paths=source_relative:. \
           --validate_out=paths=source_relative,lang=go:. \
           $(API_PROTO_FILES)

.PHONY: oswagger
# generate grpc code
oswagger:
	protoc 	$(PROTO_PATHS) \
			--openapiv2_out . \
			--openapiv2_opt logtostderr=true \
			--openapiv2_opt json_names_for_fields=false \
			$(API_PROTO_FILES)

.PHONY: ext
# generate grpc code
ext:
	protoc $(PROTO_PATHS) \
 	       --go_out=paths=source_relative:./ \
 	       --go-grpc_out=paths=source_relative:./ \
	       $(EXT_PROTO_FILES)

.PHONY: run
# run app service
run:
	go run cmd/main/main.go

.PHONY: cron
# run cron service
cron:
	go run cmd/cron/main.go $(OPTIONAL_PARAM)

.PHONY: consumer
# run consumer service
consumer:
	go run cmd/consumer/main.go

.PHONY: swagger
# run swagger service
swagger:
	swagger generate spec -m -o ./swagger.yaml
	@echo "---"
	@echo Starting to validate the swagger file
	swagger validate ./swagger.yaml --skip-warnings

.PHONY: mocks
mocks:
	mockery --dir=./internal/ --all
	mockery --dir=./external/ --all

test:
	@echo "Running tests, excluding directories: $(EXCLUDE_DIRS)"
	go test $$(go list ./... | grep -v -e '/internal/*/*_mock.go' -e '/internal/*/*_test.go' -e '/api/*/*.pb.go' -e '/api/*/*.pb.validate.go' -e '/external/*/*/*.pb.go') \
	--timeout 5m \
	-cover -coverprofile=coverage.out \
	-failfast

coverage: test
	@echo "Generating coverage report"
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: all
all:
	make grpc;
	make validate;
	make oswagger;
	protoc-go-inject-tag -input="./api/*/*.pb.go"

#docker-rune:
#	docker run \
#      -p 8080:8000 \
#      -e SERVER_PORT=8000 \
#      go-one:v1 \
#      -f /etc/app.yaml
