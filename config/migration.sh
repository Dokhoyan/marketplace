#!/bin/sh
set -eu

: "${PG_DSN:?PG_DSN is required}"
exec goose -dir "${MIGRATION_DIR:-migrations}" postgres "$PG_DSN" up
