package handlers

// Importaciones de librerías necesarias

import (
	"net/http"
	"strconv"

	"github.com/IsaacCorrales03/Mocion/backend/internal/logs"
	"github.com/IsaacCorrales03/Mocion/backend/internal/presencia"
	"github.com/IsaacCorrales03/Mocion/backend/internal/ws"
)

// NuevoPresenciaHandler crea el handler websocket que marca a un usuario como conectado a un
// debate mientras tenga esta conexión abierta (una pestaña con debate.html abierto = una
// conexión). No espera mensajes del cliente; solo importa que la conexión siga viva.
func NuevoPresenciaHandler(registro *presencia.Registro) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		debateID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

		if err != nil {
			responderError(w, http.StatusBadRequest, "ID de debate inválido")
			return
		}

		usuarioID, err := strconv.ParseInt(r.URL.Query().Get("usuario_id"), 10, 64)

		if err != nil || usuarioID == 0 {
			responderError(w, http.StatusBadRequest, "usuario_id inválido")
			return
		}

		conexion, err := ws.Upgrade(w, r)

		if err != nil {
			logs.Error("Error al establecer la conexión websocket de presencia: " + err.Error())
			return
		}

		registro.Conectar(debateID, usuarioID)
		defer registro.Desconectar(debateID, usuarioID)

		for {

			_, err := conexion.ReadMessage()

			if err != nil {
				return
			}
		}
	}
}

// NuevoConectadosHandler crea el handler que lista los IDs de usuario conectados ahora mismo a
// un debate, usado por el panel de gestión para mostrar quién ya está presente
func NuevoConectadosHandler(registro *presencia.Registro) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		debateID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

		if err != nil {
			responderError(w, http.StatusBadRequest, "ID de debate inválido")
			return
		}

		responderJSON(w, http.StatusOK, registro.Conectados(debateID))
	}
}
