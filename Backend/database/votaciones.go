package database

// Importaciones de librerías necesarias

import (
	"database/sql"
	"time"

	"github.com/IsaacCorrales03/Mocion/backend/internal/logs"
)

// Voto representa un registro de la tabla votos: a qué equipo (A o B) le da su voto el público,
// en una de las dos rondas de votación del debate
type Voto struct {
	ID        int64     `json:"id"`
	DebateID  int64     `json:"debate_id"`
	UsuarioID int64     `json:"usuario_id"`
	Ronda     int       `json:"ronda"`
	Opcion    string    `json:"opcion"`
	CreadoEn  time.Time `json:"creado_en"`
}

// ResultadoVotacion resume el conteo de votos de una ronda de votación de un debate
type ResultadoVotacion struct {
	DebateID   int64   `json:"debate_id"`
	Ronda      int     `json:"ronda"`
	EquipoA    int     `json:"equipo_a"`
	EquipoB    int     `json:"equipo_b"`
	TotalVotos int     `json:"total_votos"`
	PorcentajeA float64 `json:"porcentaje_a"`
	PorcentajeB float64 `json:"porcentaje_b"`
}

// CrearTablaVotos crea la tabla votos si no existe (esquema nuevo, para bases limpias)
func CrearTablaVotos(db *sql.DB) error {

	consulta := `
	CREATE TABLE IF NOT EXISTS votos (
		id SERIAL PRIMARY KEY,
		debate_id INTEGER NOT NULL REFERENCES debates(id),
		usuario_id INTEGER NOT NULL REFERENCES usuarios(id),
		ronda INTEGER NOT NULL DEFAULT 1,
		opcion VARCHAR(1) NOT NULL,
		creado_en TIMESTAMP NOT NULL DEFAULT NOW(),
		UNIQUE (debate_id, usuario_id, ronda)
	);`

	_, err := db.Exec(consulta)

	if err != nil {
		logs.Error("Error al crear la tabla votos: " + err.Error())
		return err
	}

	logs.Info("Tabla votos lista")

	return nil
}

// MigrarTablaVotos adapta una tabla votos de la versión anterior (a_favor/en_contra, un voto por
// debate) al nuevo esquema (a/b, un voto por ronda). Es deliberadamente conservadora: si ya
// existen votos con las opciones viejas, los dos primeros ALTER fallarían por el CHECK implícito
// del VARCHAR más corto, así que solo agrega la columna 'ronda' y reemplaza la restricción única;
// no migra datos de votaciones anteriores al nuevo formato A/B.
func MigrarTablaVotos(db *sql.DB) error {

	consulta := `
	ALTER TABLE votos ADD COLUMN IF NOT EXISTS ronda INTEGER NOT NULL DEFAULT 1;
	ALTER TABLE votos ALTER COLUMN opcion TYPE VARCHAR(20);
	ALTER TABLE votos DROP CONSTRAINT IF EXISTS votos_debate_id_usuario_id_key;
	ALTER TABLE votos DROP CONSTRAINT IF EXISTS votos_debate_id_usuario_id_ronda_key;
	ALTER TABLE votos ADD CONSTRAINT votos_debate_id_usuario_id_ronda_key UNIQUE (debate_id, usuario_id, ronda);`

	_, err := db.Exec(consulta)

	if err != nil {
		logs.Error("Error al migrar la tabla votos: " + err.Error())
		return err
	}

	logs.Info("Migración de la tabla votos lista")

	return nil
}

// RegistrarVoto guarda el voto de un usuario por un equipo (A o B), en una ronda de votación
// específica del debate. Un usuario solo puede votar una vez por ronda.
func RegistrarVoto(db *sql.DB, debateID int64, usuarioID int64, ronda int, opcion string) (*Voto, error) {

	voto := &Voto{DebateID: debateID, UsuarioID: usuarioID, Ronda: ronda, Opcion: opcion}

	consulta := `
	INSERT INTO votos (debate_id, usuario_id, ronda, opcion)
	VALUES ($1, $2, $3, $4)
	RETURNING id, creado_en;`

	err := db.QueryRow(consulta, debateID, usuarioID, ronda, opcion).Scan(&voto.ID, &voto.CreadoEn)

	if err != nil {
		logs.Error("Error al registrar el voto: " + err.Error())
		return nil, err
	}

	logs.Info("Voto registrado para el debate " + itoa(debateID))

	return voto, nil
}

// YaVoto indica si un usuario ya emitió su voto en una ronda de votación específica de un debate
func YaVoto(db *sql.DB, debateID int64, usuarioID int64, ronda int) (bool, error) {

	consulta := `SELECT EXISTS(SELECT 1 FROM votos WHERE debate_id = $1 AND usuario_id = $2 AND ronda = $3);`

	var existe bool

	err := db.QueryRow(consulta, debateID, usuarioID, ronda).Scan(&existe)

	if err != nil {
		logs.Error("Error al comprobar el voto: " + err.Error())
		return false, err
	}

	return existe, nil
}

// ObtenerResultados calcula el conteo de votos por equipo de una ronda de votación de un debate
func ObtenerResultados(db *sql.DB, debateID int64, ronda int) (*ResultadoVotacion, error) {

	resultado := &ResultadoVotacion{DebateID: debateID, Ronda: ronda}

	consulta := `
	SELECT
		COUNT(*) FILTER (WHERE opcion = $3) AS equipo_a,
		COUNT(*) FILTER (WHERE opcion = $4) AS equipo_b
	FROM votos
	WHERE debate_id = $1 AND ronda = $2;`

	err := db.QueryRow(consulta, debateID, ronda, EquipoA, EquipoB).Scan(&resultado.EquipoA, &resultado.EquipoB)

	if err != nil {
		logs.Error("Error al obtener los resultados: " + err.Error())
		return nil, err
	}

	resultado.TotalVotos = resultado.EquipoA + resultado.EquipoB

	if resultado.TotalVotos > 0 {
		resultado.PorcentajeA = float64(resultado.EquipoA) / float64(resultado.TotalVotos) * 100
		resultado.PorcentajeB = float64(resultado.EquipoB) / float64(resultado.TotalVotos) * 100
	}

	return resultado, nil
}
