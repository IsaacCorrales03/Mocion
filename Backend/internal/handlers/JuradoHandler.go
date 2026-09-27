package handlers

// Importaciones de librerías necesarias

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/IsaacCorrales03/Mocion/backend/database"
)

// asignarJuradoPeticion representa el cuerpo JSON esperado para asignar un jurado
type asignarJuradoPeticion struct {
	UsuarioID int64 `json:"usuario_id"`
}

// NuevoAsignarJuradoHandler crea el handler para agregar un miembro al jurado de un debate
func NuevoAsignarJuradoHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			responderError(w, http.StatusMethodNotAllowed, "método no permitido")
			return
		}

		debateID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

		if err != nil {
			responderError(w, http.StatusBadRequest, "ID de debate inválido")
			return
		}

		var peticion asignarJuradoPeticion

		err = json.NewDecoder(r.Body).Decode(&peticion)

		if err != nil {
			responderError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}

		if peticion.UsuarioID == 0 {
			responderError(w, http.StatusBadRequest, "el usuario es obligatorio")
			return
		}

		miembro, err := database.AsignarJurado(db, debateID, peticion.UsuarioID)

		if err != nil {
			responderError(w, http.StatusConflict, database.ErrYaEsJurado.Error())
			return
		}

		responderJSON(w, http.StatusCreated, miembro)
	}
}

// NuevoObtenerJuradoHandler crea el handler que lista los miembros del jurado de un debate
func NuevoObtenerJuradoHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		debateID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

		if err != nil {
			responderError(w, http.StatusBadRequest, "ID de debate inválido")
			return
		}

		miembros, err := database.ObtenerJurado(db, debateID)

		if err != nil {
			responderError(w, http.StatusInternalServerError, "error interno al obtener el jurado")
			return
		}

		responderJSON(w, http.StatusOK, miembros)
	}
}
