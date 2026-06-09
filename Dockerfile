FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates
WORKDIR /app

COPY ./stock-data-provider-service /app/stock-data-provider-service
COPY ./common-go-lib /app/common-go-lib
WORKDIR /app/stock-data-provider-service
RUN go mod tidy
RUN go build -o /bin/stock-data-provider-service .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/stock-data-provider-service /bin/stock-data-provider-service
COPY --from=builder /app/stock-data-provider-service/config /app/stock-data-provider-service/config
WORKDIR /app/stock-data-provider-service
ENTRYPOINT ["/bin/stock-data-provider-service"]
