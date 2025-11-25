#!/bin/sh

set -e

host_port="$1"
host="${host_port%:*}"
port="${host_port#*:}"
shift

if [ "$1" = "--" ]; then
  shift
fi

cmd="$@"

#until PGPASSWORD=$POSTGRES_PASSWORD psql -h "$host" -U "postgres" -d "mydb" -c '\q'; do
#  >&2 echo "Postgres is unavailable - sleeping"
#  sleep 1
#done

until python -c "
import psycopg2
import os
try:
  conn = psycopg2.connect(
    dbname = os.getenv('POSTGRES_DB', 'mydb'),
    user = os.getenv('POSTGRES_USER', 'postgres'),
    password = os.getenv('POSTGRES_PASSWORD', 'password'),
    host = '$host',
    port = '$port'
  )
  conn.close()
  exit(0)
except Exception as e:
  print(f'Connection failed: {e}')
  exit(1)
"; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

>&2 echo "Postgres is up - executing command"
exec $cmd