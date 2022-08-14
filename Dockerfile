FROM golang:latest as builder

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

FROM debian:stable-slim

COPY --from=builder /go/bin/ /app/
COPY configs/* /app/configs/

WORKDIR /app/

CMD ["/app/go-chat"]