FROM golang:1.14-alpine as builder

ENV GOPROXY https://goproxy.cn,direct

WORKDIR /build

COPY . .

RUN cd /build/app/service/article/cmd \
    && CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -o article ./main.go \
    && cd /build/app/service/poems/cmd \
    && CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -o poems ./main.go \
    && cd /build/app/service/webinfo/cmd \
    && CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -o webinfo ./main.go \
    && cd /build/app/interface/main/cmd \
    && CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -o main ./main.go


FROM alpine:3.10 AS final

ARG USER_NAME="micro"
ARG USER_ID=10001

WORKDIR /app

COPY --from=builder /build /app/

RUN addgroup -g ${USER_ID} -S ${USER_NAME} && adduser -u ${USER_ID} -S ${USER_NAME} -G ${USER_NAME} \
    && chown -R ${USER_NAME}:${USER_NAME} /app/

USER ${USER_NAME}

RUN touch /app/app/service/article/cmd/output.log

EXPOSE 8000 8001 8100 8101 8200 8201 8080

CMD ["/bin/sh","./entrypoint.sh"]



