package database

// Importaciones de librerías necesarias

import (
	"database/sql"
	"errors"
	"time"

	"github.com/IsaacCorrales03/Mocion/backend/internal/logs"
)

// MiembroJurado representa a alguien asignado como jurado de un debate. El veredicto ya no se
// registra como una sola opción: ahora el jurado califica a cada participante (ver
// puntuaciones.go) y el ganador se calcula automáticamente al finalizar el debate.
type MiembroJurado struct {
	ID         int64     `json:"id"`
	DebateID   int64     `json:"debate_id"`
	UsuarioID  int64     `json:"usuario_id"`
	Nombre     string    `json:"nombre"`
	AsignadoEn time.Time `json:"asignado_en"`
}

// ErrYaEsJurado se retorna cuando el usuario ya fue asignado como jurado del debate
var ErrYaEsJurado = errors.New("el usuario ya es jurado de este debate")

// CrearTablaJurado crea la tabla jurado si no existe (ya sin columnas de veredicto: la
// calificación ahora vive en puntuaciones_jurado, por participante)
func CrearTablaJurado(db *sql.DB) error {

	consulta := `
	CREATE TABLE IF NOT EXISTS jurado (
		id SERIAL PRIMARY KEY,
		debate_id INTEGER NOT NULL REFERENCES debates(id),
		usuario_id INTEGER NOT NULL REFERENCES usuarios(id),
		asignado_en TIMESTAMP NOT NULL DEFAULT NOW(),
		UNIQUE (debate_id, usuario_id)
	);`

	_, err := db.Exec(consulta)

	if err != nil {
		logs.Error("Error al crear la tabla jurado: " + err.Error())
		return err
	}

	logs.Info("Tabla jurado lista")

	return nil
}

// AsignarJurado agrega a un usuario como miembro del jurado de un debate
func AsignarJurado(db *sql.DB, debateID int64, usuarioID int64) (*MiembroJurado, error) {

	miembro := &MiembroJurado{DebateID: debateID, UsuarioID: usuarioID}

	consulta := `
	INSERT INTO jurado (debate_id, usuario_id)
	VALUES ($1, $2)
	RETURNING id, asignado_en;`

	err := db.QueryRow(consulta, debateID, usuarioID).Scan(&miembro.ID, &miembro.AsignadoEn)

	if err != nil {
		logs.Error("Error al asignar el jurado: " + err.Error())
		return nil, ErrYaEsJurado
	}

	logs.Info("Jurado asignado al debate " + itoa(debateID))

	return miembro, nil
}

// EsMiembroJurado comprueba si un usuario forma parte del jurado de un debate
func EsMiembroJurado(db *sql.DB, debateID int64, usuarioID int64) (bool, error) {

	consulta := `SELECT EXISTS(SELECT 1 FROM jurado WHERE debate_id = $1 AND usuario_id = $2);`

	var existe bool

	err := db.QueryRow(consulta, debateID, usuarioID).Scan(&existe)

	if err != nil {
		logs.Error("Error al comprobar el jurado: " + err.Error())
		return false, err
	}

	return existe, nil
}

// ObtenerJurado retorna todos los miembros del jurado de un debate, con su nombre
func ObtenerJurado(db *sql.DB, debateID int64) ([]MiembroJurado, error) {

	consulta := `
	SELECT j.id, j.debate_id, j.usuario_id, u.nombre, j.asignado_en
	FROM jurado j
	JOIN usuarios u ON u.id = j.usuario_id
	WHERE j.debate_id = $1
	ORDER BY j.asignado_en ASC;`

	filas, err := db.Query(consulta, debateID)

	if err != nil {
		logs.Error("Error al obtener el jurado: " + err.Error())
		return nil, err
	}

	defer filas.Close()

	miembros := make([]MiembroJurado, 0)

	for filas.Next() {

		var miembro MiembroJurado

		err := filas.Scan(&miembro.ID, &miembro.DebateID, &miembro.UsuarioID, &miembro.Nombre, &miembro.AsignadoEn)

		if err != nil {
			logs.Error("Error al leer un miembro del jurado: " + err.Error())
			return nil, err
		}

		miembros = append(miembros, miembro)
	}

	return miembros, nil
}
