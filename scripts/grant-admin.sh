#!/bin/sh
set -eu

if [ "$#" -ne 1 ]; then
    echo "Использование: $0 <login>" >&2
    exit 1
fi

login=$1
case "$login" in
    ''|*[!A-Za-z0-9_]*)
        echo "Логин должен содержать латинские буквы, цифры или знак подчеркивания." >&2
        exit 1
        ;;
esac
if [ "${#login}" -lt 3 ] || [ "${#login}" -gt 32 ]; then
    echo "Длина логина должна быть от 3 до 32 символов." >&2
    exit 1
fi

project_dir=$(CDPATH= cd "$(dirname "$0")/.." && pwd)

run_psql() {
    if [ -n "${PG_DSN:-}" ]; then
        if ! command -v psql >/dev/null 2>&1; then
            echo "Для подключения по PG_DSN нужен установленный psql." >&2
            exit 1
        fi
        psql -X --dbname="$PG_DSN" -v ON_ERROR_STOP=1 -qAt --set="login=$login"
    else
        docker compose --project-directory "$project_dir" exec -T pg sh -c '
            exec psql -X -v ON_ERROR_STOP=1 -qAt \
                -U "$POSTGRES_USER" -d "$POSTGRES_DB" --set="login=$1"
        ' sh "$login"
    fi
}

result=$(run_psql <<'SQL'
UPDATE users
SET role = 'admin'
WHERE lower(login) = lower(:'login') AND deleted_at IS NULL
RETURNING login;
SQL
)

if [ -z "$result" ]; then
    echo "Активный пользователь с таким логином не найден." >&2
    exit 1
fi

printf 'Пользователю %s назначена роль admin.\n' "$result"
