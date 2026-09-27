package handlers

// Importaciones de librerías necesarias

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/IsaacCorrales03/Mocion/backend/database"
)

// asignarParticipantePeticion representa el cuerpo JSON esperado para agregar a alguien a un equipo
type asignarParticipantePeticion struct {
	UsuarioID int64  `json:"usuario_id"`
	Equipo    string `json:"equipo"` // "a" o "b"
}

// NuevoAsignarParticipanteHandler crea el handler para agregar a un usuario a un equipo del debate
func NuevoAsignarParticipanteHandler(db *sql.DB) http.HandlerFunc {

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

		var peticion asignarParticipantePeticion

		err = json.NewDecoder(r.Body).Decode(&peticion)

		if err != nil {
			responderError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}

		if peticion.UsuarioID == 0 {
			responderError(w, http.StatusBadRequest, "el usuario es obligatorio")
			return
		}

		participante, err := database.AsignarParticipante(db, debateID, peticion.UsuarioID, peticion.Equipo)

		if err == database.ErrEquipoInvalido {
			responderError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err == database.ErrYaEsParticipante {
			responderError(w, http.StatusConflict, err.Error())
			return
		}

		if err != nil {
			responderError(w, http.StatusInternalServerError, "error interno al asignar el participante")
			return
		}

		responderJSON(w, http.StatusCreated, participante)
	}
}

// NuevoListarParticipantesHandler crea el handler que lista los participantes de un debate
func NuevoListarParticipantesHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		debateID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

		if err != nil {
			responderError(w, http.StatusBadRequest, "ID de debate inválido")
			return
		}

		participantes, err := database.ListarParticipantes(db, debateID)

		if err != nil {
			responderError(w, http.StatusInternalServerError, "error interno al listar participantes")
			return
		}

		responderJSON(w, http.StatusOK, participantes)
	}
}
