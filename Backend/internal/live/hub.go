package live

// Importaciones de librerías necesarias

import (
	"sync"

	"github.com/IsaacCorrales03/Mocion/backend/internal/logs"
	"github.com/IsaacCorrales03/Mocion/backend/internal/ws"
)

// sala representa un grupo de conexiones que comparten los mensajes entre sí (un debate en vivo)
type sala struct {
	mu       sync.Mutex
	clientes map[*ws.Conn]bool
}

// Hub administra varias salas identificadas por un ID de debate
type Hub struct {
	mu    sync.Mutex
	salas map[int64]*sala
}

// NuevoHub crea un Hub vacío, listo para administrar salas
func NuevoHub() *Hub {
	return &Hub{salas: make(map[int64]*sala)}
}

// obtenerOCrearSala retorna la sala asociada a un debate, creándola si aún no existe
func (h *Hub) obtenerOCrearSala(debateID int64) *sala {

	h.mu.Lock()
	defer h.mu.Unlock()

	salaExistente, existe := h.salas[debateID]

	if existe {
		return salaExistente
	}

	nuevaSala := &sala{clientes: make(map[*ws.Conn]bool)}

	h.salas[debateID] = nuevaSala

	return nuevaSala
}

// Unirse agrega una conexión a la sala de un debate
func (h *Hub) Unirse(debateID int64, conexion *ws.Conn) {

	sala := h.obtenerOCrearSala(debateID)

	sala.mu.Lock()
	sala.clientes[conexion] = true
	sala.mu.Unlock()

	logs.Info("Nueva conexión en la sala del debate " + itoa(debateID))
}

// Salir remueve una conexión de la sala de un debate
func (h *Hub) Salir(debateID int64, conexion *ws.Conn) {

	sala := h.obtenerOCrearSala(debateID)

	sala.mu.Lock()
	delete(sala.clientes, conexion)
	sala.mu.Unlock()

	conexion.Close()

	logs.Info("Conexión cerrada en la sala del debate " + itoa(debateID))
}

// Difundir envía un mensaje a todos los clientes de la sala, excepto opcionalmente a quien lo originó
func (h *Hub) Difundir(debateID int64, mensaje string, emisor *ws.Conn) {

	sala := h.obtenerOCrearSala(debateID)

	sala.mu.Lock()
	defer sala.mu.Unlock()

	for cliente := range sala.clientes {

		if cliente == emisor {
			continue
		}

		err := cliente.WriteMessage(mensaje)

		if err != nil {
			logs.Error("Error al difundir el mensaje: " + err.Error())
		}
	}
}
