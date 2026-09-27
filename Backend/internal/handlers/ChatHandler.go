package handlers

// Importaciones de librerías necesarias

import (
	"net/http"
	"strconv"

	"github.com/IsaacCorrales03/Mocion/backend/internal/live"
	"github.com/IsaacCorrales03/Mocion/backend/internal/logs"
	"github.com/IsaacCorrales03/Mocion/backend/internal/ws"
)

// NuevoChatHandler crea el handler websocket del chat en vivo, usando un Hub compartido entre debates
func NuevoChatHandler(hub *live.Hub) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		debateID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

		if err != nil {
			responderError(w, http.StatusBadRequest, "ID de debate inválido")
			return
		}

		conexion, err := ws.Upgrade(w, r)

		if err != nil {
			logs.Error("Error al establecer la conexión websocket del chat: " + err.Error())
			return
		}

		hub.Unirse(debateID, conexion)
		defer hub.Salir(debateID, conexion)

		for {

			mensaje, err := conexion.ReadMessage()

			if err != nil {
				return
			}

			hub.Difundir(debateID, mensaje, conexion)
		}
	}
}
