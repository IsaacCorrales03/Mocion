package ws

// Importaciones de librerías necesarias

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
)

// llaveMagica es la constante definida por el estándar WebSocket (RFC 6455) para calcular Sec-WebSocket-Accept
const llaveMagica = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// Conn representa una conexión WebSocket ya establecida
type Conn struct {
	rwc         net.Conn
	br          *bufio.Reader
	muEscritura sync.Mutex
}

// ErrConexionCerrada se retorna cuando el cliente cerró la conexión
var ErrConexionCerrada = errors.New("la conexión websocket fue cerrada")

// Upgrade toma una petición HTTP normal y la convierte en una conexión WebSocket
func Upgrade(w http.ResponseWriter, r *http.Request) (*Conn, error) {

	llaveCliente := r.Header.Get("Sec-WebSocket-Key")

	if llaveCliente == "" {
		http.Error(w, "se esperaba una petición de upgrade a websocket", http.StatusBadRequest)
		return nil, errors.New("falta el header Sec-WebSocket-Key")
	}

	hijacker, ok := w.(http.Hijacker)

	if !ok {
		http.Error(w, "el servidor no soporta websockets", http.StatusInternalServerError)
		return nil, errors.New("el ResponseWriter no soporta hijacking")
	}

	rwc, buf, err := hijacker.Hijack()

	if err != nil {
		return nil, err
	}

	aceptacion := calcularAceptacion(llaveCliente)

	respuesta := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + aceptacion + "\r\n\r\n"

	_, err = rwc.Write([]byte(respuesta))

	if err != nil {
		rwc.Close()
		return nil, err
	}

	return &Conn{rwc: rwc, br: buf.Reader}, nil
}

// calcularAceptacion genera el valor de Sec-WebSocket-Accept a partir de la llave del cliente
func calcularAceptacion(llaveCliente string) string {

	suma := sha1.Sum([]byte(llaveCliente + llaveMagica))

	return base64.StdEncoding.EncodeToString(suma[:])
}

// ReadMessage lee mensajes de texto del cliente, ignorando de forma transparente los frames de
// control (ping/pong) que no son parte de la conversación: si el cliente manda un ping, se le
// responde con un pong automáticamente y se sigue esperando el próximo mensaje real. Antes estos
// frames de control se devolvían tal cual como si fueran un mensaje de texto, y terminaban
// reenviándose al resto de la sala como si fueran señalización WebRTC válida.
func (c *Conn) ReadMessage() (string, error) {

	for {

		opcode, datos, err := c.leerFrame()

		if err != nil {
			return "", err
		}

		switch opcode {

		case 0x8: // frame de cierre enviado por el cliente
			return "", ErrConexionCerrada

		case 0x9: // ping: se responde con un pong y se sigue esperando el próximo mensaje
			if err := c.writeFrame(0xA, datos); err != nil {
				return "", err
			}
			continue

		case 0xA: // pong: no requiere respuesta, se ignora
			continue

		default: // 0x1 (texto) o 0x2 (binario): es un mensaje real
			return string(datos), nil
		}
	}
}

// leerFrame lee un único frame WebSocket del cliente y devuelve su opcode y payload sin interpretar
func (c *Conn) leerFrame() (byte, []byte, error) {

	primerByte, err := c.br.ReadByte()

	if err != nil {
		return 0, nil, err
	}

	opcode := primerByte & 0x0F

	segundoByte, err := c.br.ReadByte()

	if err != nil {
		return 0, nil, err
	}

	enmascarado := segundoByte&0x80 != 0
	longitud := int64(segundoByte & 0x7F)

	if longitud == 126 {

		var extendida uint16

		err = binary.Read(c.br, binary.BigEndian, &extendida)

		if err != nil {
			return 0, nil, err
		}

		longitud = int64(extendida)

	} else if longitud == 127 {

		var extendida uint64

		err = binary.Read(c.br, binary.BigEndian, &extendida)

		if err != nil {
			return 0, nil, err
		}

		longitud = int64(extendida)
	}

	var mascara [4]byte

	if enmascarado {

		_, err = io.ReadFull(c.br, mascara[:])

		if err != nil {
			return 0, nil, err
		}
	}

	datos := make([]byte, longitud)

	_, err = io.ReadFull(c.br, datos)

	if err != nil {
		return 0, nil, err
	}

	if enmascarado {
		for i := range datos {
			datos[i] ^= mascara[i%4]
		}
	}

	return opcode, datos, nil
}

// WriteMessage envía un mensaje de texto al cliente (los frames del servidor van sin máscara)
func (c *Conn) WriteMessage(mensaje string) error {
	return c.writeFrame(0x1, []byte(mensaje))
}

// writeFrame arma y envía un frame WebSocket con el opcode indicado. Protegido con un mutex porque
// varias goroutines pueden querer escribirle a la misma conexión al mismo tiempo (p. ej. dos
// personas hablando en el chat a la vez, o varios candidatos ICE llegando juntos durante la
// señalización de streaming): sin este candado, dos escrituras concurrentes podían intercalar sus
// bytes y corromper el frame, lo que el navegador simplemente descarta o interpreta como un cierre
// de la conexión — la causa más probable de que la transmisión se cortara sin ningún error visible.
func (c *Conn) writeFrame(opcode byte, datos []byte) error {

	longitud := len(datos)

	var encabezado []byte

	switch {

	case longitud <= 125:
		encabezado = []byte{0x80 | opcode, byte(longitud)}

	case longitud <= 65535:
		encabezado = []byte{0x80 | opcode, 126, byte(longitud >> 8), byte(longitud)}

	default:
		encabezado = make([]byte, 10)
		encabezado[0] = 0x80 | opcode
		encabezado[1] = 127
		binary.BigEndian.PutUint64(encabezado[2:], uint64(longitud))
	}

	c.muEscritura.Lock()
	defer c.muEscritura.Unlock()

	_, err := c.rwc.Write(append(encabezado, datos...))

	return err
}

// Close cierra la conexión WebSocket
func (c *Conn) Close() error {
	return c.rwc.Close()
}
