package live

// MensajeChat representa un mensaje enviado dentro del chat en vivo de un debate
type MensajeChat struct {
	Tipo      string `json:"tipo"`
	UsuarioID int64  `json:"usuario_id"`
	Nombre    string `json:"nombre"`
	Contenido string `json:"contenido"`
}

// MensajeSenal representa un mensaje de señalización WebRTC (oferta, respuesta o candidato ICE)
// que se retransmite tal cual entre el transmisor y los espectadores de un debate
type MensajeSenal struct {
	Tipo      string `json:"tipo"`
	UsuarioID int64  `json:"usuario_id"`
	Contenido string `json:"contenido"`
}
