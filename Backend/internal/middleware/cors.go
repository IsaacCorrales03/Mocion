package middleware

// Importaciones de librerías necesarias

import (
	"net/http"
	"os"
	"strings"
)

// CORS envuelve un http.Handler y agrega las cabeceras necesarias para permitir
// que el frontend (servido desde otro origen, p. ej. http://localhost:3000)
// pueda llamar a esta API. Los orígenes permitidos se toman de la variable de
// entorno ORIGENES_PERMITIDOS (separados por coma); si no está definida, se
// permite cualquier origen ("*"), útil para desarrollo.
func CORS(siguiente http.Handler) http.Handler {

	origenesPermitidos := strings.Split(os.Getenv("ORIGENES_PERMITIDOS"), ",")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		origen := r.Header.Get("Origin")

		if origenPermitido(origen, origenesPermitidos) {
			w.Header().Set("Access-Control-Allow-Origin", origen)
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		siguiente.ServeHTTP(w, r)
	})
}

// origenPermitido determina si el origen de la petición está en la lista blanca.
// Una lista vacía o que contenga "*" permite cualquier origen (modo desarrollo).
func origenPermitido(origen string, permitidos []string) bool {

	if origen == "" {
		return false
	}

	for _, permitido := range permitidos {

		permitido = strings.TrimSpace(permitido)

		if permitido == "*" || permitido == "" || permitido == origen {
			return true
		}
	}

	return false
}
