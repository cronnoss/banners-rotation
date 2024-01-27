API_BIN := "./bin/banner"
DOCKER_IMG="banner:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2

lint: install-lint-deps
	golangci-lint run ./...

.PHONY: build run build-img run-img version test lint

generate:
	rm -rf internal/server/pb
	mkdir -p internal/server/pb

	protoc \
        --proto_path=./api/ \
        --go_opt=paths=source_relative \
        --go-grpc_opt=paths=source_relative \
        --go_out=./internal/server/pb \
        --go-grpc_out=./internal/server/pb \
        api/*.proto

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/banner/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

up:
	docker-compose -f docker-compose.yaml up --build -d ;\
	docker-compose up -d

down:
	docker-compose down

test:
	go clean -testcache;
	go test -v -race -count 100 ./internal/...

integration-tests:
		set -e ;\
    	docker-compose -f docker-compose.test.yaml up --build -d ;\
    	test_status_code=0 ;\
    	docker-compose -f docker-compose.test.yaml run integration_tests go test ./test/integration_test.go || test_status_code=$$? ;\
    	docker-compose -f docker-compose.test.yaml down ;\
    	echo $$test_status_code ;\
    	exit $$test_status_code ;


build:
	go build -v -o $(API_BIN) -ldflags "$(LDFLAGS)" ./cmd/banner

run: build
	$(API_BIN) -config ./configs/banner_config.yaml

build-debug:
	go build -gcflags="all=-N -l" -o $(API_BIN) -ldflags "$(LDFLAGS)" ./cmd/banner

run-debug: build-debug
	$(API_BIN) -config ./configs/banner_config.yaml

run-postgres:
	docker run -d --rm --name composepostgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e PGDATA=/var/lib/postgresql/data/pgdata -v psqldata:/var/lib/postgresql/data -p 5432:5432 postgres:latest

migrate-up:
	goose -dir migrations postgres "host=localhost user=postgres password=postgres dbname=postgres sslmode=disable" up

migrate-down:
	goose -dir migrations postgres "host=localhost user=postgres password=postgres dbname=postgres sslmode=disable" down

migrate-down-to-0001:
	goose -dir migrations postgres "host=localhost user=postgres password=postgres dbname=postgres sslmode=disable" down-to 0001

migrate-status:
	goose -dir migrations postgres "host=localhost user=postgres password=postgres dbname=postgres sslmode=disable" status

migrate-reset:
	goose -dir migrations postgres "host=localhost user=postgres password=postgres dbname=postgres sslmode=disable" reset

run-rabbitmq:
	docker run -d --rm --name rabbitmq -e RABBITMQ_DEFAULT_USER=guest -e RABBITMQ_DEFAULT_PASS=guest -p 15672:15672 -p 5672:5672 rabbitmq:3-management

stop-rabbitmq:
	docker stop rabbitmq
