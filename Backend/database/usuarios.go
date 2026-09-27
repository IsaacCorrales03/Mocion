package database

// Importaciones de librerías necesarias

import (
	"database/sql"
	"errors"
	"time"

	"github.com/IsaacCorrales03/Mocion/backend/internal/logs"
)

// Usuario representa un registro de la tabla usuarios
type Usuario struct {
	ID           int64
	Nombre       string
	Email        string
	PasswordHash string
	Salt         string
	CreadoEn     time.Time
}

// ErrUsuarioNoEncontrado se retorna cuando no existe un usuario con el email indicado
var ErrUsuarioNoEncontrado = errors.New("usuario no encontrado")

// ErrCorreoYaRegistrado se retorna cuando el correo ya existe en la base de datos
var ErrCorreoYaRegistrado = errors.New("el correo ya está registrado")

// CrearTablaUsuarios crea la tabla usuarios si no existe
func CrearTablaUsuarios(db *sql.DB) error {

	consulta := `
	CREATE TABLE IF NOT EXISTS usuarios (
		id SERIAL PRIMARY KEY,
		nombre VARCHAR(100) NOT NULL,
		email VARCHAR(150) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		salt VARCHAR(255) NOT NULL,
		creado_en TIMESTAMP NOT NULL DEFAULT NOW()
	);`

	_, err := db.Exec(consulta)

	if err != nil {
		logs.Error("Error al crear la tabla usuarios: " + err.Error())
		return err
	}

	logs.Info("Tabla usuarios lista")

	return nil
}

// CrearUsuario inserta un nuevo usuario en la base de datos
func CrearUsuario(db *sql.DB, nombre string, email string, passwordHash string, salt string) (*Usuario, error) {

	usuario := &Usuario{
		Nombre:       nombre,
		Email:        email,
		PasswordHash: passwordHash,
		Salt:         salt,
	}

	consulta := `
	INSERT INTO usuarios (nombre, email, password_hash, salt)
	VALUES ($1, $2, $3, $4)
	RETURNING id, creado_en;`

	err := db.QueryRow(consulta, nombre, email, passwordHash, salt).Scan(&usuario.ID, &usuario.CreadoEn)

	if err != nil {
		logs.Error("Error al crear el usuario: " + err.Error())
		return nil, err
	}

	logs.Info("Usuario creado: " + email)

	return usuario, nil
}

// ObtenerUsuarioPorEmail busca un usuario a partir de su correo
func ObtenerUsuarioPorEmail(db *sql.DB, email string) (*Usuario, error) {

	usuario := &Usuario{}

	consulta := `
	SELECT id, nombre, email, password_hash, salt, creado_en
	FROM usuarios
	WHERE email = $1;`

	fila := db.QueryRow(consulta, email)

	err := fila.Scan(&usuario.ID, &usuario.Nombre, &usuario.Email, &usuario.PasswordHash, &usuario.Salt, &usuario.CreadoEn)

	if err == sql.ErrNoRows {
		return nil, ErrUsuarioNoEncontrado
	}

	if err != nil {
		logs.Error("Error al obtener el usuario por email: " + err.Error())
		return nil, err
	}

	return usuario, nil
}

// ExisteCorreo indica si ya existe un usuario registrado con ese correo
func ExisteCorreo(db *sql.DB, email string) (bool, error) {

	_, err := ObtenerUsuarioPorEmail(db, email)

	if err == ErrUsuarioNoEncontrado {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

// UsuarioPublico expone solo los campos de un usuario seguros de mostrar a otros usuarios
type UsuarioPublico struct {
	ID     int64  `json:"id"`
	Nombre string `json:"nombre"`
	Email  string `json:"email"`
}

// BuscarUsuarios busca usuarios cuyo nombre o correo coincida parcialmente con el texto dado,
// usado por quien organiza un debate para encontrar a quién asignar como participante o jurado
func BuscarUsuarios(db *sql.DB, texto string) ([]UsuarioPublico, error) {

	consulta := `
	SELECT id, nombre, email
	FROM usuarios
	WHERE nombre ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%'
	ORDER BY nombre ASC
	LIMIT 10;`

	filas, err := db.Query(consulta, texto)

	if err != nil {
		logs.Error("Error al buscar usuarios: " + err.Error())
		return nil, err
	}

	defer filas.Close()

	usuarios := make([]UsuarioPublico, 0)

	for filas.Next() {

		var usuario UsuarioPublico

		err := filas.Scan(&usuario.ID, &usuario.Nombre, &usuario.Email)

		if err != nil {
			logs.Error("Error al leer un usuario: " + err.Error())
			return nil, err
		}

		usuarios = append(usuarios, usuario)
	}

	return usuarios, nil
}
