package database

// Importaciones de librerías necesarias

import (
	"database/sql"
	"errors"
	"time"

	"github.com/IsaacCorrales03/Mocion/backend/internal/logs"
)

// PuntuacionJurado representa la nota (1 a 10) que un miembro del jurado le da a un participante
type PuntuacionJurado struct {
	ID              int64     `json:"id"`
	DebateID        int64     `json:"debate_id"`
	JuradoUsuarioID int64     `json:"jurado_usuario_id"`
	ParticipanteID  int64     `json:"participante_usuario_id"`
	Puntuacion      int       `json:"puntuacion"`
	CreadoEn        time.Time `json:"creado_en"`
}

// ErrPuntuacionInvalida se retorna cuando la nota está fuera del rango 1-10
var ErrPuntuacionInvalida = errors.New("la puntuación debe estar entre 1 y 10")

// CrearTablaPuntuaciones crea la tabla puntuaciones_jurado si no existe
func CrearTablaPuntuaciones(db *sql.DB) error {

	consulta := `
	CREATE TABLE IF NOT EXISTS puntuaciones_jurado (
		id SERIAL PRIMARY KEY,
		debate_id INTEGER NOT NULL REFERENCES debates(id),
		jurado_usuario_id INTEGER NOT NULL REFERENCES usuarios(id),
		participante_usuario_id INTEGER NOT NULL REFERENCES usuarios(id),
		puntuacion INTEGER NOT NULL CHECK (puntuacion BETWEEN 1 AND 10),
		creado_en TIMESTAMP NOT NULL DEFAULT NOW(),
		UNIQUE (debate_id, jurado_usuario_id, participante_usuario_id)
	);`

	_, err := db.Exec(consulta)

	if err != nil {
		logs.Error("Error al crear la tabla puntuaciones_jurado: " + err.Error())
		return err
	}

	logs.Info("Tabla puntuaciones_jurado lista")

	return nil
}

// RegistrarPuntuacion guarda (o actualiza, si el jurado corrige antes del cierre) la nota que un
// jurado le da a un participante
func RegistrarPuntuacion(db *sql.DB, debateID int64, juradoUsuarioID int64, participanteID int64, puntuacion int) (*PuntuacionJurado, error) {

	if puntuacion < 1 || puntuacion > 10 {
		return nil, ErrPuntuacionInvalida
	}

	p := &PuntuacionJurado{DebateID: debateID, JuradoUsuarioID: juradoUsuarioID, ParticipanteID: participanteID, Puntuacion: puntuacion}

	consulta := `
	INSERT INTO puntuaciones_jurado (debate_id, jurado_usuario_id, participante_usuario_id, puntuacion)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (debate_id, jurado_usuario_id, participante_usuario_id)
	DO UPDATE SET puntuacion = EXCLUDED.puntuacion
	RETURNING id, creado_en;`

	err := db.QueryRow(consulta, debateID, juradoUsuarioID, participanteID, puntuacion).Scan(&p.ID, &p.CreadoEn)

	if err != nil {
		logs.Error("Error al registrar la puntuación: " + err.Error())
		return nil, err
	}

	return p, nil
}

// ObtenerPuntuaciones retorna todas las notas registradas en un debate
func ObtenerPuntuaciones(db *sql.DB, debateID int64) ([]PuntuacionJurado, error) {

	consulta := `
	SELECT id, debate_id, jurado_usuario_id, participante_usuario_id, puntuacion, creado_en
	FROM puntuaciones_jurado
	WHERE debate_id = $1;`

	filas, err := db.Query(consulta, debateID)

	if err != nil {
		logs.Error("Error al obtener las puntuaciones: " + err.Error())
		return nil, err
	}

	defer filas.Close()

	puntuaciones := make([]PuntuacionJurado, 0)

	for filas.Next() {

		var p PuntuacionJurado

		err := filas.Scan(&p.ID, &p.DebateID, &p.JuradoUsuarioID, &p.ParticipanteID, &p.Puntuacion, &p.CreadoEn)

		if err != nil {
			logs.Error("Error al leer una puntuación: " + err.Error())
			return nil, err
		}

		puntuaciones = append(puntuaciones, p)
	}

	return puntuaciones, nil
}

// PromedioJuradoPorEquipo calcula el promedio (0-10) de las notas del jurado para cada equipo de
// un debate, uniendo puntuaciones_jurado con participantes para saber a qué equipo pertenece cada
// participante calificado
func PromedioJuradoPorEquipo(db *sql.DB, debateID int64) (promedioA float64, promedioB float64, err error) {

	consulta := `
	SELECT
		COALESCE(AVG(pj.puntuacion) FILTER (WHERE p.equipo = $2), 0),
		COALESCE(AVG(pj.puntuacion) FILTER (WHERE p.equipo = $3), 0)
	FROM puntuaciones_jurado pj
	JOIN participantes p ON p.debate_id = pj.debate_id AND p.usuario_id = pj.participante_usuario_id
	WHERE pj.debate_id = $1;`

	err = db.QueryRow(consulta, debateID, EquipoA, EquipoB).Scan(&promedioA, &promedioB)

	if err != nil {
		logs.Error("Error al calcular el promedio del jurado: " + err.Error())
		return 0, 0, err
	}

	return promedioA, promedioB, nil
}
