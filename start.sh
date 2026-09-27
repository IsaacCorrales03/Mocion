#!/bin/bash
# start.sh — levanta Postgres (docker-compose.yml), el backend (Go) y el
# frontend (`serve`) juntos.
# Uso: ./start.sh   (desde la raíz del proyecto, junto a Backend/, Frontend/
# y docker-compose.yml)

set -e

BACKEND_PID=""

cleanup() {
  echo ""
  echo "Cerrando..."
  if [ -n "$BACKEND_PID" ] && kill -0 "$BACKEND_PID" 2>/dev/null; then
    kill "$BACKEND_PID"
  fi
  exit 0
}
trap cleanup INT TERM

echo "==> Levantando Postgres (docker compose)..."
docker compose up -d

# Detecta el nombre del servicio de Postgres dentro del compose
SERVICIO_DB=$(docker compose config --services | grep -Ei 'postgres|psql|^db$' | head -n1)
if [ -z "$SERVICIO_DB" ]; then
  SERVICIO_DB=$(docker compose config --services | head -n1)
fi

echo "==> Esperando a que Postgres (${SERVICIO_DB}) acepte conexiones..."
until docker compose exec -T "$SERVICIO_DB" pg_isready -U Administrador -d mocion > /dev/null 2>&1; do
  sleep 1
done
echo "    Postgres listo."

echo "==> Levantando backend (Go)..."
cd Backend
go run main.go &
BACKEND_PID=$!
cd ..

sleep 1

echo "==> Levantando frontend (npx serve)..."
cd Frontend
npx serve

cleanup
