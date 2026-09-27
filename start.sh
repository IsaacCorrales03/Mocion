
#!/bin/bash
# start-backend.sh — levanta Postgres (docker compose) + compila y arranca
# el backend. Pensado para correr en tu propio servidor (no en un PaaS que
# no te deje manejar Docker vos mismo).
#
# Variables de entorno que el backend necesita (poné un Backend/.env o
# exportalas antes de correr esto):
#   URL_BASE_DE_DATOS    -> postgres://Administrador:mocion2026@localhost:5432/mocion?sslmode=disable
#   ORIGENES_PERMITIDOS  -> URL del frontend, cuando ya la tengas
#   PORT                 -> opcional, 8080 si no la seteás

set -e

cd "$(dirname "$0")"

echo "==> Levantando Postgres (docker compose)..."
docker compose up -d

SERVICIO_DB=$(docker compose config --services | grep -Ei 'postgres|psql|^db$' | head -n1)
if [ -z "$SERVICIO_DB" ]; then
  SERVICIO_DB=$(docker compose config --services | head -n1)
fi

echo "==> Esperando a que Postgres (${SERVICIO_DB}) acepte conexiones..."
until docker compose exec -T "$SERVICIO_DB" pg_isready -U Administrador -d mocion > /dev/null 2>&1; do
  sleep 1
done
echo "    Postgres listo."

cd Backend

echo "==> Compilando backend..."
go build -o mocion-backend main.go

echo "==> Arrancando backend..."
exec ./mocion-backend
