package handlers

// Importaciones de librerías necesarias

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/IsaacCorrales03/Mocion/backend/database"
)

// puntuarPeticion representa el cuerpo JSON esperado para que un jurado califique a un participante
type puntuarPeticion struct {
	JuradoUsuarioID int64 `json:"jurado_usuario_id"`
	ParticipanteID  int64 `json:"participante_usuario_id"`
	Puntuacion      int   `json:"puntuacion"`
}

// NuevoPuntuarHandler crea el handler para que un miembro del jurado califique (1-10) a un
// participante. Se puede llamar varias veces por participante mientras el debate no haya
// finalizado: cada llamada actualiza la nota anterior de ese jurado para ese participante.
func NuevoPuntuarHandler(db *sql.DB) http.HandlerFunc {

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

		var peticion puntuarPeticion

		err = json.NewDecoder(r.Body).Decode(&peticion)

		if err != nil {
			responderError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}

		esJurado, err := database.EsMiembroJurado(db, debateID, peticion.JuradoUsuarioID)

		if err != nil {
			responderError(w, http.StatusInternalServerError, "error interno al comprobar el jurado")
			return
		}

		if !esJurado {
			responderError(w, http.StatusForbidden, "el usuario no es jurado de este debate")
			return
		}

		puntuacion, err := database.RegistrarPuntuacion(db, debateID, peticion.JuradoUsuarioID, peticion.ParticipanteID, peticion.Puntuacion)

		if err == database.ErrPuntuacionInvalida {
			responderError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err != nil {
			responderError(w, http.StatusInternalServerError, "error interno al registrar la puntuación")
			return
		}

		responderJSON(w, http.StatusOK, puntuacion)
	}
}

// NuevoObtenerPuntuacionesHandler crea el handler que lista todas las notas registradas en un debate
func NuevoObtenerPuntuacionesHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		debateID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

		if err != nil {
			responderError(w, http.StatusBadRequest, "ID de debate inválido")
			return
		}

		puntuaciones, err := database.ObtenerPuntuaciones(db, debateID)

		if err != nil {
			responderError(w, http.StatusInternalServerError, "error interno al obtener las puntuaciones")
			return
		}

		responderJSON(w, http.StatusOK, puntuaciones)
	}
}
