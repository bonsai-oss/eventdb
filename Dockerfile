FROM golang:bullseye AS builder
WORKDIR /build
ENV GO111MODULE=on
ENV CGO_ENABLE=no
COPY . .
RUN go get -u ./...
RUN go build -o eventdb cmd/eventdb/main.go

FROM alpine:edge
RUN mkdir /app
RUN adduser -D user && chown -R user /app
COPY --from=builder /build/eventdb /app
EXPOSE 8080
USER user
ENTRYPOINT ["/app/eventdb"]

