FROM golang:latest as builder

WORKDIR /go/src/github.com/sqrthree/progressbar201X

RUN go get -u github.com/golang/dep/cmd/dep

COPY Gopkg.toml Gopkg.lock ./

RUN dep ensure -vendor-only

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -installsuffix cgo -o app ./cmd/progressbar201X/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /go/src/github.com/sqrthree/progressbar201X/app .

EXPOSE 3000

CMD ["./app"]
