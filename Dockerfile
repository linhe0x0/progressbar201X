FROM golang:latest as builder

WORKDIR /go/src/github.com/sqrthree/progressbar201X

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && chmod +x /usr/local/bin/dep

COPY Gopkg.toml Gopkg.lock ./

RUN dep ensure -vendor-only

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -installsuffix cgo -o app ./cmd/progressbar201X/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /go/src/github.com/sqrthree/progressbar201X/app .

EXPOSE 3000

ENTRYPOINT /root/app
# CMD ["./app"]
