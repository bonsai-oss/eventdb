FROM golang:bullseye AS builder
WORKDIR /build
ENV GO111MODULE=on
ENV CGO_ENABLE=no
COPY . .
RUN go get -u ./...
RUN go build -o eventdb cmd/eventdb/main.go

FROM alpine:edge
WORKDIR /app
COPY --from=builder /build/eventdb /app
RUN ls -lah /app
EXPOSE 8080
ENTRYPOINT ["/app/eventdb"]

