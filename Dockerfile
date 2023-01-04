FROM golang:bullseye AS builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o eventdb -trimpath -ldflags '-s -w' cmd/eventdb/main.go

FROM alpine:edge
RUN mkdir /app
RUN adduser -D user && chown -R user /app
COPY --from=builder /build/eventdb /app
EXPOSE 8080
USER user
ENTRYPOINT ["/app/eventdb"]
