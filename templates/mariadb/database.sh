#!/bin/bash
mysql -h {{.DatabaseHostname}} -u {{.DatabaseAdminUsername}} -P 3306 -e "CREATE DATABASE IF NOT EXISTS {{.DatabaseName}}; GRANT ALL PRIVILEGES ON {{.DatabaseName}}.* TO '{{.DatabaseName}}'@'localhost' IDENTIFIED BY '$DATABASE_PASSWORD';GRANT ALL PRIVILEGES ON {{.DatabaseName}}.* TO '{{.DatabaseName}}'@'%' IDENTIFIED BY '$DATABASE_PASSWORD';"
