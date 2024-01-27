FROM golang:1.21

ENV CODE_DIR /go/src

# Set destination for COPY
WORKDIR ${CODE_DIR}

# Caching layers with modules
COPY ./go.mod .
COPY ./go.sum .

# Download Go modules
RUN go mod download

COPY . ${CODE_DIR}

ENV CGO_ENABLED=0

CMD go test -v -tags integration ./test