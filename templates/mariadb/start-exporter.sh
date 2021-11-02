#!/bin/bash
export DATA_SOURCE_NAME="root:${MARIADB_ROOT_PASSWORD}@(localhost:3306)/"
exec mysqld_exporter
