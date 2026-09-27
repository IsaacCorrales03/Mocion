package handlers

// Importaciones de librerías necesarias

import (
	"database/sql"
	"net/http"

	"github.com/IsaacCorrales03/Mocion/backend/database"
)

// NuevoBuscarUsuariosHandler crea el handler que busca usuarios por nombre o correo (parcial),
// usado por quien organiza un debate para encontrar a quién agregar como participante o jurado
func NuevoBuscarUsuariosHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		texto := r.URL.Query().Get("q")

		if len(texto) < 2 {
			responderJSON(w, http.StatusOK, []database.UsuarioPublico{})
			return
		}

		usuarios, err := database.BuscarUsuarios(db, texto)

		if err != nil {
			responderError(w, http.StatusInternalServerError, "error interno al buscar usuarios")
			return
		}

		responderJSON(w, http.StatusOK, usuarios)
	}
}
