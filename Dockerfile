FROM golang

WORKDIR /app

COPY cmd cmd
COPY internal internal
COPY go.mod go.mod
COPY go.sum go.sum
COPY vendor vendor

RUN CGO_ENABLED=0 go build -mod vendor -o proxy ./cmd/proxy

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=0 /app/proxy .

CMD ["/root/proxy"]  
