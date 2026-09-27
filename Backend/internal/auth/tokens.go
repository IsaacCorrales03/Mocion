package auth

// Importaciones de librerías necesarias

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

// duracionSesion indica cuánto tiempo dura un token antes de expirar
const duracionSesion = 24 * time.Hour

// sesion representa una sesión activa asociada a un token
type sesion struct {
	UsuarioID int64
	ExpiraEn  time.Time
}

// almacenSesiones guarda las sesiones activas en memoria, protegidas con un mutex
var almacenSesiones = struct {
	sync.RWMutex
	datos map[string]sesion
}{datos: make(map[string]sesion)}

// ErrTokenInvalido se retorna cuando el token no existe o ya expiró
var ErrTokenInvalido = errors.New("token inválido o expirado")

// generarToken crea un token aleatorio de 32 bytes en formato hexadecimal
func generarToken() (string, error) {

	bytesAleatorios := make([]byte, 32)

	_, err := rand.Read(bytesAleatorios)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytesAleatorios), nil
}

// CrearSesion genera un nuevo token para un usuario y lo guarda en el almacén
func CrearSesion(usuarioID int64) (string, error) {

	token, err := generarToken()

	if err != nil {
		return "", err
	}

	almacenSesiones.Lock()
	almacenSesiones.datos[token] = sesion{
		UsuarioID: usuarioID,
		ExpiraEn:  time.Now().Add(duracionSesion),
	}
	almacenSesiones.Unlock()

	return token, nil
}

// ValidarToken comprueba si un token es válido y retorna el ID del usuario asociado
func ValidarToken(token string) (int64, error) {

	almacenSesiones.RLock()
	sesionEncontrada, existe := almacenSesiones.datos[token]
	almacenSesiones.RUnlock()

	if !existe {
		return 0, ErrTokenInvalido
	}

	if time.Now().After(sesionEncontrada.ExpiraEn) {
		almacenSesiones.Lock()
		delete(almacenSesiones.datos, token)
		almacenSesiones.Unlock()

		return 0, ErrTokenInvalido
	}

	return sesionEncontrada.UsuarioID, nil
}

// CerrarSesion elimina un token del almacén, invalidándolo
func CerrarSesion(token string) {

	almacenSesiones.Lock()
	delete(almacenSesiones.datos, token)
	almacenSesiones.Unlock()
}
