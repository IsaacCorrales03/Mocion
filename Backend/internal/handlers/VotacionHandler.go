package handlers

// Importaciones de librerías necesarias

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/IsaacCorrales03/Mocion/backend/database"
)

// votarPeticion representa el cuerpo JSON esperado para emitir un voto por un equipo
type votarPeticion struct {
	UsuarioID int64  `json:"usuario_id"`
	Opcion    string `json:"opcion"` // "a" o "b"
}

// NuevoVotarHandler crea el handler para registrar el voto de un usuario por un equipo. La ronda
// de votación (1ª o 2ª) se toma de la fase actual del debate, no del cuerpo de la petición — el
// público solo puede votar durante las ventanas de votación que abre el motor de fases.
func NuevoVotarHandler(db *sql.DB) http.HandlerFunc {

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

		var peticion votarPeticion

		err = json.NewDecoder(r.Body).Decode(&peticion)

		if err != nil {
			responderError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}

		if peticion.Opcion != database.EquipoA && peticion.Opcion != database.EquipoB {
			responderError(w, http.StatusBadRequest, "la opción debe ser 'a' o 'b'")
			return
		}

		debate, err := database.ObtenerDebatePorID(db, debateID)

		if err == database.ErrDebateNoEncontrado {
			responderError(w, http.StatusNotFound, err.Error())
			return
		}

		if err != nil {
			responderError(w, http.StatusInternalServerError, "error interno al obtener el debate")
			return
		}

		var ronda int

		switch debate.Fase {
		case database.FaseVotacion1:
			ronda = 1
		case database.FaseVotacion2:
			ronda = 2
		default:
			responderError(w, http.StatusConflict, "no hay una votación abierta en este momento")
			return
		}

		yaVoto, err := database.YaVoto(db, debateID, peticion.UsuarioID, ronda)

		if err != nil {
			responderError(w, http.StatusInternalServerError, "error interno al comprobar el voto")
			return
		}

		if yaVoto {
			responderError(w, http.StatusConflict, "el usuario ya votó en esta ronda")
			return
		}

		voto, err := database.RegistrarVoto(db, debateID, peticion.UsuarioID, ronda, peticion.Opcion)

		if err != nil {
			responderError(w, http.StatusInternalServerError, "error interno al registrar el voto")
			return
		}

		responderJSON(w, http.StatusCreated, voto)
	}
}

// NuevoResultadosHandler crea el handler que retorna el conteo de votos de una ronda de votación
// del debate (?ronda=1 por defecto, o ?ronda=2 para la segunda votación)
func NuevoResultadosHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		debateID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

		if err != nil {
			responderError(w, http.StatusBadRequest, "ID de debate inválido")
			return
		}

		ronda := 1

		if valor := r.URL.Query().Get("ronda"); valor != "" {
			if n, err := strconv.Atoi(valor); err == nil {
				ronda = n
			}
		}

		resultados, err := database.ObtenerResultados(db, debateID, ronda)

		if err != nil {
			responderError(w, http.StatusInternalServerError, "error interno al obtener los resultados")
			return
		}

		responderJSON(w, http.StatusOK, resultados)
	}
}
