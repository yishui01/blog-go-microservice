FROM golang:1.15 as builder

ARG SERVICE_NAME="article_service"
ARG APP_DIR="/app"
ARG RUN_DIR=${APP_DIR}/${SERVICE_NAME}

WORKDIR /build

RUN addgroup -g 1001 -S ${SERVICE_NAME} && adduser -u 1001 -S ${SERVICE_NAME} -G ${SERVICE_NAME}

ENV GOPROXY https://goproxy.cn,direct

COPY go.mod .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -o ${SERVICE_NAME} ./cmd/

FROM alpine:3.10 AS final

WORKDIR /app

COPY --from=builder /build/${SERVICE_NAME} /app/

USER ${SERVICE_NAME}


ENTRYPOINT ["/bin/sh","${RUN_DIR}"]







