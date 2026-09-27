#!/bin/bash
# start.sh — levanta el backend (Go) y el frontend (servido con `serve`) juntos.
# Uso: ./start.sh   (desde la raíz del proyecto, donde están las carpetas Backend/ y Frontend/)

set -e

# Al salir (Ctrl+C, error, etc.) mata también el proceso del backend
cleanup() {
  echo ""
  echo "Cerrando..."
  if [ -n "$BACKEND_PID" ] && kill -0 "$BACKEND_PID" 2>/dev/null; then
    kill "$BACKEND_PID"
  fi
  exit 0
}
trap cleanup INT TERM

echo "==> Levantando backend (Go)..."
cd Backend
go run main.go &
BACKEND_PID=$!
cd ..

# pequeña espera para que el backend esté escuchando antes de abrir el frontend
sleep 1

echo "==> Levantando frontend (npx serve)..."
cd Frontend
npx serve

# si `serve` termina, apagamos también el backend
cleanup
