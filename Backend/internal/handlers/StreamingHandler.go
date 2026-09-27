package handlers

// Importaciones de librerías necesarias

import (
	"net/http"
	"strconv"

	"github.com/IsaacCorrales03/Mocion/backend/internal/live"
	"github.com/IsaacCorrales03/Mocion/backend/internal/logs"
	"github.com/IsaacCorrales03/Mocion/backend/internal/ws"
)

// NuevoStreamingHandler crea el handler websocket usado para retransmitir la señalización WebRTC
// (ofertas, respuestas y candidatos ICE) entre el transmisor de un debate y sus espectadores.
// El video en sí viaja directamente entre navegadores vía WebRTC; este handler solo conecta a las partes.
func NuevoStreamingHandler(hub *live.Hub) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		debateID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

		if err != nil {
			responderError(w, http.StatusBadRequest, "ID de debate inválido")
			return
		}

		conexion, err := ws.Upgrade(w, r)

		if err != nil {
			logs.Error("Error al establecer la conexión websocket de streaming: " + err.Error())
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
