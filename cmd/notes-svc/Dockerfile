FROM golang:1.24.5-alpine as BUILDER
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY internal/pkg ./internal/pkg
COPY internal/notes-svc ./internal/notes-svc
COPY cmd/notes-svc ./cmd/notes-svc

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/app ./cmd/notes-svc

FROM alpine
WORKDIR /app
COPY --from=builder /bin/app /app
CMD ["./app"]