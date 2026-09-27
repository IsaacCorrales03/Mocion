package presencia

// Importaciones de librerías necesarias

import "sync"

// Registro lleva la cuenta, en memoria, de qué usuarios tienen abierta ahora mismo la página
// de un debate (una conexión websocket de presencia por pestaña abierta). Se usa para exigir
// que todos los participantes y jurados asignados estén realmente conectados antes de arrancar.
type Registro struct {
	mu      sync.Mutex
	activos map[int64]map[int64]int // debateID -> usuarioID -> cantidad de pestañas abiertas
}

// Nuevo crea un registro de presencia vacío
func Nuevo() *Registro {
	return &Registro{activos: make(map[int64]map[int64]int)}
}

// Conectar suma una conexión activa de un usuario a un debate (si abre varias pestañas, cuenta
// cada una, para que cerrar una sola no lo marque como desconectado)
func (r *Registro) Conectar(debateID int64, usuarioID int64) {

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.activos[debateID] == nil {
		r.activos[debateID] = make(map[int64]int)
	}

	r.activos[debateID][usuarioID]++
}

// Desconectar resta una conexión; cuando llega a cero, el usuario deja de contar como conectado
func (r *Registro) Desconectar(debateID int64, usuarioID int64) {

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.activos[debateID] == nil {
		return
	}

	r.activos[debateID][usuarioID]--

	if r.activos[debateID][usuarioID] <= 0 {
		delete(r.activos[debateID], usuarioID)
	}

	if len(r.activos[debateID]) == 0 {
		delete(r.activos, debateID)
	}
}

// Conectados retorna los IDs de usuario actualmente conectados a un debate
func (r *Registro) Conectados(debateID int64) []int64 {

	r.mu.Lock()
	defer r.mu.Unlock()

	lista := make([]int64, 0)

	for id := range r.activos[debateID] {
		lista = append(lista, id)
	}

	return lista
}

// EstaConectado comprueba si un usuario puntual está conectado a un debate en este momento
func (r *Registro) EstaConectado(debateID int64, usuarioID int64) bool {

	r.mu.Lock()
	defer r.mu.Unlock()

	return r.activos[debateID][usuarioID] > 0
}
