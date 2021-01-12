#!/bin/bash
set -e

password="${MARIADB_ROOT_PASSWORD:-}"
if [[ -f "${MARIADB_ROOT_PASSWORD_FILE:-}" ]]; then
    password=$(cat "$MARIADB_ROOT_PASSWORD_FILE")
fi

mysqladmin status -uroot -p"${password}"
