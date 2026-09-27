package database

// Importaciones de librerías necesarias

import (
	"database/sql"
	"errors"
	"time"

	"github.com/IsaacCorrales03/Mocion/backend/internal/logs"
)

// Estados posibles de un debate. "abierto" es antes de comenzar (se arman los equipos y el
// jurado); "transmitiendo" cubre todas las fases en vivo (presentación, moción, votaciones y
// las rondas de debate); "finalizado" es el estado terminal, con resultado ya calculado.
const (
	EstadoAbierto       = "abierto"
	EstadoTransmitiendo = "transmitiendo"
	EstadoFinalizado    = "finalizado"
)

// Fases dentro del estado "transmitiendo". Fuera de ese estado, Fase queda en FaseNinguna.
const (
	FaseNinguna      = ""
	FasePresentacion = "presentacion"
	FaseMocion       = "mocion"
	FaseVotacion1    = "votacion_1"
	FaseDebate       = "debate"
	FaseVotacion2    = "votacion_2"
	FaseFinalizado   = "finalizado"
)

// Equipos de un debate
const (
	EquipoA = "a"
	EquipoB = "b"
)

// Tipos de turno dentro de la fase de debate (y de presentación/moción)
const (
	TurnoPresentacion   = "presentacion"
	TurnoMocion         = "mocion"
	TurnoArgumento      = "argumento"
	TurnoReplica        = "replica"
	TurnoContrarreplica = "contrarreplica"
)

// Debate representa un registro de la tabla debates
type Debate struct {
	ID                   int64      `json:"id"`
	Titulo               string     `json:"titulo"`
	Descripcion          string     `json:"descripcion"`
	CreadorID            int64      `json:"creador_id"`
	Estado               string     `json:"estado"`
	Rondas               int        `json:"rondas"`
	ParticipantesXEquipo int        `json:"participantes_por_equipo"`
	Fase                 string     `json:"fase"`
	RondaActual          int        `json:"ronda_actual"`
	FaseTerminaEn        *time.Time `json:"fase_termina_en,omitempty"`
	TurnoUsuarioID       *int64     `json:"turno_usuario_id,omitempty"`
	TurnoEquipo          *string    `json:"turno_equipo,omitempty"`
	TurnoTipo            *string    `json:"turno_tipo,omitempty"`
	PuntuacionA          *float64   `json:"puntuacion_a,omitempty"`
	PuntuacionB          *float64   `json:"puntuacion_b,omitempty"`
	GanadorEquipo        *string    `json:"ganador_equipo,omitempty"`
	CreadoEn             time.Time  `json:"creado_en"`
	IniciadoEn           *time.Time `json:"iniciado_en,omitempty"`
	FinalizadoEn         *time.Time `json:"finalizado_en,omitempty"`
}

// ErrDebateNoEncontrado se retorna cuando no existe un debate con el ID indicado
var ErrDebateNoEncontrado = errors.New("debate no encontrado")

// CrearTablaDebates crea la tabla debates si no existe (esquema nuevo, para bases limpias)
func CrearTablaDebates(db *sql.DB) error {

	consulta := `
	CREATE TABLE IF NOT EXISTS debates (
		id SERIAL PRIMARY KEY,
		titulo VARCHAR(150) NOT NULL,
		descripcion TEXT NOT NULL DEFAULT '',
		creador_id INTEGER NOT NULL REFERENCES usuarios(id),
		estado VARCHAR(20) NOT NULL DEFAULT 'abierto',
		rondas INTEGER NOT NULL DEFAULT 3,
		participantes_por_equipo INTEGER NOT NULL DEFAULT 2,
		fase VARCHAR(20) NOT NULL DEFAULT '',
		ronda_actual INTEGER NOT NULL DEFAULT 0,
		fase_termina_en TIMESTAMP,
		turno_usuario_id INTEGER REFERENCES usuarios(id),
		turno_equipo VARCHAR(1),
		turno_tipo VARCHAR(20),
		puntuacion_a DOUBLE PRECISION,
		puntuacion_b DOUBLE PRECISION,
		ganador_equipo VARCHAR(1),
		creado_en TIMESTAMP NOT NULL DEFAULT NOW(),
		iniciado_en TIMESTAMP,
		finalizado_en TIMESTAMP
	);`

	_, err := db.Exec(consulta)

	if err != nil {
		logs.Error("Error al crear la tabla debates: " + err.Error())
		return err
	}

	logs.Info("Tabla debates lista")

	return nil
}

// MigrarTablaDebates agrega las columnas del motor de fases a una tabla debates que ya
// existía de una versión anterior (CREATE TABLE IF NOT EXISTS no las habría agregado)
func MigrarTablaDebates(db *sql.DB) error {

	consulta := `
	ALTER TABLE debates
		ADD COLUMN IF NOT EXISTS rondas INTEGER NOT NULL DEFAULT 3,
		ADD COLUMN IF NOT EXISTS participantes_por_equipo INTEGER NOT NULL DEFAULT 2,
		ADD COLUMN IF NOT EXISTS fase VARCHAR(20) NOT NULL DEFAULT '',
		ADD COLUMN IF NOT EXISTS ronda_actual INTEGER NOT NULL DEFAULT 0,
		ADD COLUMN IF NOT EXISTS fase_termina_en TIMESTAMP,
		ADD COLUMN IF NOT EXISTS turno_usuario_id INTEGER REFERENCES usuarios(id),
		ADD COLUMN IF NOT EXISTS turno_equipo VARCHAR(1),
		ADD COLUMN IF NOT EXISTS turno_tipo VARCHAR(20),
		ADD COLUMN IF NOT EXISTS puntuacion_a DOUBLE PRECISION,
		ADD COLUMN IF NOT EXISTS puntuacion_b DOUBLE PRECISION,
		ADD COLUMN IF NOT EXISTS ganador_equipo VARCHAR(1);`

	_, err := db.Exec(consulta)

	if err != nil {
		logs.Error("Error al migrar la tabla debates: " + err.Error())
		return err
	}

	logs.Info("Migración de la tabla debates lista")

	return nil
}

// columnasDebate es la lista de columnas usada por todos los SELECT de esta tabla, para
// mantener el orden sincronizado con Scan
const columnasDebate = `
	id, titulo, descripcion, creador_id, estado, rondas, participantes_por_equipo,
	fase, ronda_actual, fase_termina_en, turno_usuario_id, turno_equipo, turno_tipo,
	puntuacion_a, puntuacion_b, ganador_equipo, creado_en, iniciado_en, finalizado_en`

func escanearDebate(fila interface{ Scan(...interface{}) error }) (*Debate, error) {

	debate := &Debate{}

	err := fila.Scan(
		&debate.ID, &debate.Titulo, &debate.Descripcion, &debate.CreadorID, &debate.Estado,
		&debate.Rondas, &debate.ParticipantesXEquipo, &debate.Fase, &debate.RondaActual,
		&debate.FaseTerminaEn, &debate.TurnoUsuarioID, &debate.TurnoEquipo, &debate.TurnoTipo,
		&debate.PuntuacionA, &debate.PuntuacionB, &debate.GanadorEquipo,
		&debate.CreadoEn, &debate.IniciadoEn, &debate.FinalizadoEn,
	)

	return debate, err
}

// CrearDebate inserta un nuevo debate en estado "abierto"
func CrearDebate(db *sql.DB, titulo string, descripcion string, creadorID int64, rondas int, participantesXEquipo int) (*Debate, error) {

	if rondas < 1 || rondas > 5 {
		rondas = 3
	}

	if participantesXEquipo < 1 {
		participantesXEquipo = 2
	}

	consulta := `
	INSERT INTO debates (titulo, descripcion, creador_id, estado, rondas, participantes_por_equipo)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING ` + columnasDebate + `;`

	fila := db.QueryRow(consulta, titulo, descripcion, creadorID, EstadoAbierto, rondas, participantesXEquipo)

	debate, err := escanearDebate(fila)

	if err != nil {
		logs.Error("Error al crear el debate: " + err.Error())
		return nil, err
	}

	logs.Info("Debate creado: " + titulo)

	return debate, nil
}

// ObtenerDebatePorID busca un debate a partir de su ID
func ObtenerDebatePorID(db *sql.DB, id int64) (*Debate, error) {

	consulta := `SELECT ` + columnasDebate + ` FROM debates WHERE id = $1;`

	debate, err := escanearDebate(db.QueryRow(consulta, id))

	if err == sql.ErrNoRows {
		return nil, ErrDebateNoEncontrado
	}

	if err != nil {
		logs.Error("Error al obtener el debate: " + err.Error())
		return nil, err
	}

	return debate, nil
}

// ListarDebatesEnVivo retorna todos los debates que están transmitiendo en este momento
func ListarDebatesEnVivo(db *sql.DB) ([]Debate, error) {
	return listarDebatesPorEstado(db, EstadoTransmitiendo)
}

// ListarDebatesProgramados retorna todos los debates todavía en estado "abierto"
func ListarDebatesProgramados(db *sql.DB) ([]Debate, error) {
	return listarDebatesPorEstado(db, EstadoAbierto)
}

// listarDebatesPorEstado es una función auxiliar para listar debates según su estado
func listarDebatesPorEstado(db *sql.DB, estado string) ([]Debate, error) {

	consulta := `SELECT ` + columnasDebate + ` FROM debates WHERE estado = $1 ORDER BY creado_en DESC;`

	filas, err := db.Query(consulta, estado)

	if err != nil {
		logs.Error("Error al listar debates: " + err.Error())
		return nil, err
	}

	defer filas.Close()

	debates := make([]Debate, 0)

	for filas.Next() {

		debate, err := escanearDebate(filas)

		if err != nil {
			logs.Error("Error al leer un debate: " + err.Error())
			return nil, err
		}

		debates = append(debates, *debate)
	}

	return debates, nil
}

// IniciarDebate cambia el estado de un debate de "abierto" a "transmitiendo". La validación de
// si hay participantes/jurado suficientes ocurre en el handler, no aquí.
func IniciarDebate(db *sql.DB, id int64) error {

	consulta := `
	UPDATE debates
	SET estado = $1, iniciado_en = NOW()
	WHERE id = $2 AND estado = $3;`

	resultado, err := db.Exec(consulta, EstadoTransmitiendo, id, EstadoAbierto)

	if err != nil {
		logs.Error("Error al iniciar el debate: " + err.Error())
		return err
	}

	filasAfectadas, _ := resultado.RowsAffected()

	if filasAfectadas == 0 {
		return errors.New("el debate no existe o no está en estado abierto")
	}

	logs.Info("Debate iniciado, ID: " + itoa(id))

	return nil
}

// FinalizarDebate cambia el estado de un debate a "finalizado" sin tocar el resultado
// (usado para un cierre forzado/manual; el cierre normal pasa por FinalizarConResultado)
func FinalizarDebate(db *sql.DB, id int64) error {

	consulta := `
	UPDATE debates
	SET estado = $1, fase = $2, finalizado_en = NOW()
	WHERE id = $3 AND estado != $1;`

	resultado, err := db.Exec(consulta, EstadoFinalizado, FaseFinalizado, id)

	if err != nil {
		logs.Error("Error al finalizar el debate: " + err.Error())
		return err
	}

	filasAfectadas, _ := resultado.RowsAffected()

	if filasAfectadas == 0 {
		return errors.New("el debate no existe o ya estaba finalizado")
	}

	logs.Info("Debate finalizado, ID: " + itoa(id))

	return nil
}

// FinalizarConResultado cierra un debate guardando el resultado final calculado por el motor
// de fases: puntuación ponderada (jurado 70% + público 30%) de cada equipo y el equipo ganador
func FinalizarConResultado(db *sql.DB, id int64, puntuacionA float64, puntuacionB float64, ganador string) error {

	consulta := `
	UPDATE debates
	SET estado = $1, fase = $2, finalizado_en = NOW(),
	    puntuacion_a = $3, puntuacion_b = $4, ganador_equipo = $5
	WHERE id = $6;`

	_, err := db.Exec(consulta, EstadoFinalizado, FaseFinalizado, puntuacionA, puntuacionB, ganador, id)

	if err != nil {
		logs.Error("Error al finalizar el debate con resultado: " + err.Error())
		return err
	}

	logs.Info("Debate finalizado con resultado, ID: " + itoa(id))

	return nil
}

// ActualizarFase guarda el avance del motor de fases: en qué fase y ronda va el debate, de quién
// es el turno (si aplica) y cuándo termina esa fase, para que cualquier cliente que consulte el
// debate (o se reconecte) pueda reconstruir dónde va la transmisión
func ActualizarFase(db *sql.DB, debateID int64, fase string, rondaActual int, turnoUsuarioID *int64, turnoEquipo *string, turnoTipo *string, terminaEn *time.Time) error {

	consulta := `
	UPDATE debates
	SET fase = $1, ronda_actual = $2, turno_usuario_id = $3, turno_equipo = $4,
	    turno_tipo = $5, fase_termina_en = $6
	WHERE id = $7;`

	_, err := db.Exec(consulta, fase, rondaActual, turnoUsuarioID, turnoEquipo, turnoTipo, terminaEn, debateID)

	if err != nil {
		logs.Error("Error al actualizar la fase del debate: " + err.Error())
		return err
	}

	return nil
}
