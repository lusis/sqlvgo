#!/bin/bash
set -eu

echo "Creating db if it doesn't exist"
mysql -u"$SQL_USERNAME" -P"$DB_PORT" -p"$SQL_PASSWORD" -h "$DB_HOST" -e "create database if not exists $DB_NAME";
echo "Creating tables"
mysql -h "$DB_HOST" -u"$SQL_USERNAME" -P"$DB_PORT" -p"$SQL_PASSWORD" "$DB_NAME" < /sql/create.sql