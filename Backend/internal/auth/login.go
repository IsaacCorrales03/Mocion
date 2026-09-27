package auth

// Importaciones de librerías necesarias

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/IsaacCorrales03/Mocion/backend/database"
	"github.com/IsaacCorrales03/Mocion/backend/internal/logs"
)

// ErrCredencialesInvalidas se retorna cuando el correo no existe o la contraseña no coincide
var ErrCredencialesInvalidas = errors.New("correo o contraseña incorrectos")

// IniciarSesion valida las credenciales de un usuario y, si son correctas, genera un token de sesión
func IniciarSesion(db *sql.DB, email string, password string) (token string, usuario *database.Usuario, err error) {

	email = strings.TrimSpace(strings.ToLower(email))

	usuarioEncontrado, err := database.ObtenerUsuarioPorEmail(db, email)

	if err == database.ErrUsuarioNoEncontrado {
		return "", nil, ErrCredencialesInvalidas
	}

	if err != nil {
		return "", nil, err
	}

	if !VerificarPassword(password, usuarioEncontrado.PasswordHash, usuarioEncontrado.Salt) {
		return "", nil, ErrCredencialesInvalidas
	}

	token, err = CrearSesion(usuarioEncontrado.ID)

	if err != nil {
		logs.Error("Error al crear la sesión: " + err.Error())
		return "", nil, err
	}

	logs.Info("Inicio de sesión exitoso: " + email)

	return token, usuarioEncontrado, nil
}
