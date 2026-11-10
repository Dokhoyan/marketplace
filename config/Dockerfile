FROM golang:1.27-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/http-server ./cmd/http-server

FROM alpine:3.24
RUN apk add --no-cache ca-certificates
WORKDIR /app
RUN mkdir logs
COPY --from=build /bin/http-server /app/http-server
EXPOSE 8080
ENTRYPOINT ["/app/http-server"]
