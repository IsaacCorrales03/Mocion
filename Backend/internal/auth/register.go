package auth

// Importaciones de librerías necesarias

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/IsaacCorrales03/Mocion/backend/database"
	"github.com/IsaacCorrales03/Mocion/backend/internal/logs"
)

// ErrDatosInvalidos se retorna cuando falta información obligatoria o tiene un formato incorrecto
var ErrDatosInvalidos = errors.New("datos de registro inválidos")

// ErrPasswordCorta se retorna cuando la contraseña no cumple con el largo mínimo
var ErrPasswordCorta = errors.New("la contraseña debe tener al menos 8 caracteres")

// RegistrarUsuario valida los datos, crea el hash de la contraseña y guarda al usuario en la base de datos
func RegistrarUsuario(db *sql.DB, nombre string, email string, password string) (*database.Usuario, error) {

	nombre = strings.TrimSpace(nombre)
	email = strings.TrimSpace(strings.ToLower(email))

	if nombre == "" || email == "" || password == "" {
		return nil, ErrDatosInvalidos
	}

	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return nil, ErrDatosInvalidos
	}

	if len(password) < 8 {
		return nil, ErrPasswordCorta
	}

	existe, err := database.ExisteCorreo(db, email)

	if err != nil {
		return nil, err
	}

	if existe {
		return nil, database.ErrCorreoYaRegistrado
	}

	hash, salt, err := HashPassword(password)

	if err != nil {
		logs.Error("Error al generar el hash de la contraseña: " + err.Error())
		return nil, err
	}

	usuario, err := database.CrearUsuario(db, nombre, email, hash, salt)

	if err != nil {
		return nil, err
	}

	return usuario, nil
}
