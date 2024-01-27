FROM golang:1.21 as build

ENV BIN_FILE /opt/banner/banner-app
ENV CODE_DIR /go/src/

# Set destination for COPY
WORKDIR ${CODE_DIR}

# Caching layers with modules
COPY ../../go.mod .
COPY ../../go.sum .

# Download Go modules
RUN go mod download

COPY build/banner ${CODE_DIR}

# Build a static Go binary (without C API dependencies),
# otherwise it will not work in the alpine image.
ARG LDFLAGS
COPY . /go/src/
RUN CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} cmd/banner/*

# The output is a thin image
FROM alpine:3.19

LABEL ORGANIZATION="OTUS Online Education"
LABEL SERVICE="banner"
LABEL MAINTAINERS="student@otus.ru"

ENV BIN_FILE "/opt/banner/banner-app"
COPY --from=build ${BIN_FILE} ${BIN_FILE}

ENV CONFIG_FILE /etc/banner/banner_config.yaml
COPY ./configs/banner_config.yaml ${CONFIG_FILE}

ENV CONFIG_MIGRATION /etc/migrations
COPY ./migrations ${CONFIG_MIGRATION}

CMD ${BIN_FILE} -config ${CONFIG_FILE}
