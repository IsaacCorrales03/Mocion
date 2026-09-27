package database

// Importaciones de librerías necesarias

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/IsaacCorrales03/Mocion/backend/internal/logs"
)

// Participante representa a alguien asignado a un equipo (A o B) de un debate
type Participante struct {
	ID        int64     `json:"id"`
	DebateID  int64     `json:"debate_id"`
	UsuarioID int64     `json:"usuario_id"`
	Nombre    string    `json:"nombre"`
	Equipo    string    `json:"equipo"`
	Orden     int       `json:"orden"`
	CreadoEn  time.Time `json:"creado_en"`
}

// ErrYaEsParticipante se retorna cuando el usuario ya fue asignado a algún equipo del debate
var ErrYaEsParticipante = errors.New("el usuario ya es participante de este debate")

// ErrEquipoInvalido se retorna cuando el equipo no es "a" ni "b"
var ErrEquipoInvalido = errors.New("el equipo debe ser 'a' o 'b'")

// CrearTablaParticipantes crea la tabla participantes si no existe
func CrearTablaParticipantes(db *sql.DB) error {

	consulta := `
	CREATE TABLE IF NOT EXISTS participantes (
		id SERIAL PRIMARY KEY,
		debate_id INTEGER NOT NULL REFERENCES debates(id),
		usuario_id INTEGER NOT NULL REFERENCES usuarios(id),
		equipo VARCHAR(1) NOT NULL,
		orden INTEGER NOT NULL DEFAULT 0,
		creado_en TIMESTAMP NOT NULL DEFAULT NOW(),
		UNIQUE (debate_id, usuario_id)
	);`

	_, err := db.Exec(consulta)

	if err != nil {
		logs.Error("Error al crear la tabla participantes: " + err.Error())
		return err
	}

	logs.Info("Tabla participantes lista")

	return nil
}

// AsignarParticipante agrega a un usuario a un equipo de un debate. El "orden" (para la rotación
// de turnos) se asigna automáticamente según cuántos ya hay en ese equipo.
func AsignarParticipante(db *sql.DB, debateID int64, usuarioID int64, equipo string) (*Participante, error) {

	if equipo != EquipoA && equipo != EquipoB {
		return nil, ErrEquipoInvalido
	}

	consulta := `
	INSERT INTO participantes (debate_id, usuario_id, equipo, orden)
	VALUES ($1, $2, $3, (SELECT COUNT(*) FROM participantes WHERE debate_id = $1 AND equipo = $4))
	RETURNING id, creado_en, orden;`

	participante := &Participante{DebateID: debateID, UsuarioID: usuarioID, Equipo: equipo}

	err := db.QueryRow(consulta, debateID, usuarioID, equipo, equipo).Scan(&participante.ID, &participante.CreadoEn, &participante.Orden)

	if err != nil {

		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, ErrYaEsParticipante
		}

		logs.Error("Error al asignar el participante: " + err.Error())
		return nil, err
	}

	logs.Info("Participante asignado al debate " + itoa(debateID))

	return participante, nil
}

// ListarParticipantes retorna todos los participantes de un debate, con su nombre, ordenados por
// equipo y por orden de rotación
func ListarParticipantes(db *sql.DB, debateID int64) ([]Participante, error) {

	consulta := `
	SELECT p.id, p.debate_id, p.usuario_id, u.nombre, p.equipo, p.orden, p.creado_en
	FROM participantes p
	JOIN usuarios u ON u.id = p.usuario_id
	WHERE p.debate_id = $1
	ORDER BY p.equipo ASC, p.orden ASC;`

	filas, err := db.Query(consulta, debateID)

	if err != nil {
		logs.Error("Error al listar participantes: " + err.Error())
		return nil, err
	}

	defer filas.Close()

	participantes := make([]Participante, 0)

	for filas.Next() {

		var p Participante

		err := filas.Scan(&p.ID, &p.DebateID, &p.UsuarioID, &p.Nombre, &p.Equipo, &p.Orden, &p.CreadoEn)

		if err != nil {
			logs.Error("Error al leer un participante: " + err.Error())
			return nil, err
		}

		participantes = append(participantes, p)
	}

	return participantes, nil
}

// ContarPorEquipo retorna cuántos participantes hay en cada equipo de un debate
func ContarPorEquipo(db *sql.DB, debateID int64) (equipoA int, equipoB int, err error) {

	consulta := `
	SELECT
		COUNT(*) FILTER (WHERE equipo = $2),
		COUNT(*) FILTER (WHERE equipo = $3)
	FROM participantes
	WHERE debate_id = $1;`

	err = db.QueryRow(consulta, debateID, EquipoA, EquipoB).Scan(&equipoA, &equipoB)

	if err != nil {
		logs.Error("Error al contar participantes por equipo: " + err.Error())
		return 0, 0, err
	}

	return equipoA, equipoB, nil
}
