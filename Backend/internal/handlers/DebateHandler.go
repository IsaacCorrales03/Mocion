package handlers

// Importaciones de librerías necesarias

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/IsaacCorrales03/Mocion/backend/database"
	"github.com/IsaacCorrales03/Mocion/backend/internal/motor"
	"github.com/IsaacCorrales03/Mocion/backend/internal/presencia"
)

// crearDebatePeticion representa el cuerpo JSON esperado para crear un debate
type crearDebatePeticion struct {
	Titulo               string `json:"titulo"`
	Descripcion          string `json:"descripcion"`
	CreadorID            int64  `json:"creador_id"`
	Rondas               int    `json:"rondas"`
	ParticipantesXEquipo int    `json:"participantes_por_equipo"`
}

// iniciarDebatePeticion representa el cuerpo JSON opcional al iniciar un debate: si los equipos
// no están completos, hay que confirmar explícitamente con "forzar": true para arrancar igual
type iniciarDebatePeticion struct {
	Forzar bool `json:"forzar"`
}

// NuevoCrearDebateHandler crea el handler para registrar un nuevo debate en estado "abierto"
func NuevoCrearDebateHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			responderError(w, http.StatusMethodNotAllowed, "método no permitido")
			return
		}

		var peticion crearDebatePeticion

		err := json.NewDecoder(r.Body).Decode(&peticion)

		if err != nil {
			responderError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}

		if peticion.Titulo == "" || peticion.CreadorID == 0 {
			responderError(w, http.StatusBadRequest, "el título y el creador son obligatorios")
			return
		}

		debate, err := database.CrearDebate(db, peticion.Titulo, peticion.Descripcion, peticion.CreadorID, peticion.Rondas, peticion.ParticipantesXEquipo)

		if err != nil {
			responderError(w, http.StatusInternalServerError, "error interno al crear el debate")
			return
		}

		responderJSON(w, http.StatusCreated, debate)
	}
}

// NuevoObtenerDebateHandler crea el handler para consultar un debate por su ID
func NuevoObtenerDebateHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

		if err != nil {
			responderError(w, http.StatusBadRequest, "ID de debate inválido")
			return
		}

		debate, err := database.ObtenerDebatePorID(db, id)

		if err == database.ErrDebateNoEncontrado {
			responderError(w, http.StatusNotFound, err.Error())
			return
		}

		if err != nil {
			responderError(w, http.StatusInternalServerError, "error interno al obtener el debate")
			return
		}

		responderJSON(w, http.StatusOK, debate)
	}
}

// NuevoExplorarDebatesHandler crea el handler que lista los debates que están transmitiendo ahora
func NuevoExplorarDebatesHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		debates, err := database.ListarDebatesEnVivo(db)

		if err != nil {
			responderError(w, http.StatusInternalServerError, "error interno al explorar debates")
			return
		}

		responderJSON(w, http.StatusOK, debates)
	}
}

// NuevoDebatesProgramadosHandler crea el handler que lista los debates aún en estado "abierto"
func NuevoDebatesProgramadosHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		debates, err := database.ListarDebatesProgramados(db)

		if err != nil {
			responderError(w, http.StatusInternalServerError, "error interno al listar debates abiertos")
			return
		}

		responderJSON(w, http.StatusOK, debates)
	}
}

// NuevoIniciarDebateHandler crea el handler que arranca la transmisión de un debate: valida que
// los equipos estén completos y que todos los asignados (participantes y jurado) estén realmente
// conectados en este momento — a menos que se fuerce explícitamente — y dispara el motor de fases.
func NuevoIniciarDebateHandler(db *sql.DB, m *motor.Motor, registro *presencia.Registro) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			responderError(w, http.StatusMethodNotAllowed, "método no permitido")
			return
		}

		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

		if err != nil {
			responderError(w, http.StatusBadRequest, "ID de debate inválido")
			return
		}

		var peticion iniciarDebatePeticion
		_ = json.NewDecoder(r.Body).Decode(&peticion) // cuerpo opcional, se ignora si viene vacío

		debate, err := database.ObtenerDebatePorID(db, id)

		if err == database.ErrDebateNoEncontrado {
			responderError(w, http.StatusNotFound, err.Error())
			return
		}

		if err != nil {
			responderError(w, http.StatusInternalServerError, "error interno al obtener el debate")
			return
		}

		if !peticion.Forzar {

			problemas := make([]string, 0)

			cantidadA, cantidadB, err := database.ContarPorEquipo(db, id)

			if err != nil {
				responderError(w, http.StatusInternalServerError, "error interno al contar participantes")
				return
			}

			if cantidadA < debate.ParticipantesXEquipo {
				problemas = append(problemas, fmt.Sprintf("faltan %d en el equipo A", debate.ParticipantesXEquipo-cantidadA))
			}

			if cantidadB < debate.ParticipantesXEquipo {
				problemas = append(problemas, fmt.Sprintf("faltan %d en el equipo B", debate.ParticipantesXEquipo-cantidadB))
			}

			participantes, err := database.ListarParticipantes(db, id)

			if err != nil {
				responderError(w, http.StatusInternalServerError, "error interno al listar participantes")
				return
			}

			jurado, err := database.ObtenerJurado(db, id)

			if err != nil {
				responderError(w, http.StatusInternalServerError, "error interno al obtener el jurado")
				return
			}

			desconectados := make([]string, 0)

			for _, p := range participantes {
				if !registro.EstaConectado(id, p.UsuarioID) {
					desconectados = append(desconectados, p.Nombre)
				}
			}

			for _, j := range jurado {
				if !registro.EstaConectado(id, j.UsuarioID) {
					desconectados = append(desconectados, j.Nombre)
				}
			}

			if len(desconectados) > 0 {
				problemas = append(problemas, "no están conectados ahora mismo: "+strings.Join(desconectados, ", "))
			}

			if len(problemas) > 0 {
				responderJSON(w, http.StatusConflict, map[string]interface{}{
					"error":                 strings.Join(problemas, "; "),
					"requiere_confirmacion": true,
				})
				return
			}
		}

		err = database.IniciarDebate(db, id)

		if err != nil {
			responderError(w, http.StatusConflict, err.Error())
			return
		}

		m.Iniciar(id)

		responderJSON(w, http.StatusOK, map[string]string{"estado": database.EstadoTransmitiendo})
	}
}

// NuevoFinalizarDebateHandler crea el handler que corta manualmente un debate en transmisión
// (cierre forzado, sin el cálculo de resultado que hace el motor al terminar solo)
func NuevoFinalizarDebateHandler(db *sql.DB, m *motor.Motor) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			responderError(w, http.StatusMethodNotAllowed, "método no permitido")
			return
		}

		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

		if err != nil {
			responderError(w, http.StatusBadRequest, "ID de debate inválido")
			return
		}

		m.Cancelar(id)

		err = database.FinalizarDebate(db, id)

		if err != nil {
			responderError(w, http.StatusConflict, err.Error())
			return
		}

		responderJSON(w, http.StatusOK, map[string]string{"estado": database.EstadoFinalizado})
	}
}
