FROM golang:1.16-alpine3.13 as builder

WORKDIR $GOPATH/src/wechat
COPY . .

RUN apk add --no-cache git && set -x && \
    go mod init && go get -d -v
RUN CGO_ENABLED=0 GOOS=linux go build -o /server server.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /s5 s5.go

FROM alpine:latest

WORKDIR /
COPY --from=builder /server . 
COPY --from=builder /s5 . 

ADD entrypoint.sh /entrypoint.sh


RUN  chmod +x /server /s5  && chmod 777 /entrypoint.sh
ENTRYPOINT  /entrypoint.sh 

EXPOSE 8080
