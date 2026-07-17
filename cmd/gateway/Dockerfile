FROM golang:1.26.4-alpine as BUILDER
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY internal/pkg ./internal/pkg
COPY internal/gateway ./internal/gateway
COPY cmd/gateway ./cmd/gateway

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/app ./cmd/gateway

FROM alpine
WORKDIR /app
COPY --from=builder /bin/app /app
CMD ["./app"]