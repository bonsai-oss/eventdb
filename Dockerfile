FROM golang AS builder
RUN go build -o eventdb cmd/eventdb/main.go

FROM alpine:edge
RUN mkdir /app
RUN adduser -D user && chown -R user /app
COPY --from=builder eventdb /app
USER user
ENTRYPOINT ["/app/eventdb"]

