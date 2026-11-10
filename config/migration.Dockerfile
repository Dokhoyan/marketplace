FROM golang:1.27-alpine AS build
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

FROM alpine:3.24
RUN apk add --no-cache ca-certificates
COPY --from=build /go/bin/goose /usr/local/bin/goose
WORKDIR /app
COPY migrations/ migrations/
COPY config/migration.sh migration.sh
ENTRYPOINT ["sh", "migration.sh"]
