package motor

// Importaciones de librerías necesarias

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/IsaacCorrales03/Mocion/backend/database"
	"github.com/IsaacCorrales03/Mocion/backend/internal/live"
	"github.com/IsaacCorrales03/Mocion/backend/internal/logs"
)

// Duraciones de cada tipo de turno e intervalo. Los valores marcados como "asumido" no fueron
// especificados exactamente por el organizador y se dejaron en un valor razonable — ajustables
// aquí mismo si hace falta otro tiempo.
const (
	duracionPresentacion   = 15 * time.Second
	duracionMocion         = 60 * time.Second // asumido: no se especificó tiempo para plantear la moción
	duracionVotacion1      = 30 * time.Second
	duracionArgumento      = 90 * time.Second
	duracionReplica        = 60 * time.Second
	duracionContrarreplica = 30 * time.Second
	duracionVotacion2      = 90 * time.Second // asumido: ventana para 2ª votación pública + puntuación del jurado

	pausaAntesDeArgumento      = 60 * time.Second
	pausaAntesDeReplica        = 60 * time.Second
	pausaAntesDeContrarreplica = 30 * time.Second
)

// mensajeFase es lo que se retransmite por el socket de streaming a todos los presentes en un
// debate cada vez que el motor avanza de fase o de turno, para que la interfaz se actualice sin
// tener que hacer polling
type mensajeFase struct {
	Tipo           string     `json:"tipo"` // siempre "fase"
	Fase           string     `json:"fase"`
	RondaActual    int        `json:"ronda_actual"`
	TurnoUsuarioID *int64     `json:"turno_usuario_id,omitempty"`
	TurnoEquipo    *string    `json:"turno_equipo,omitempty"`
	TurnoTipo      *string    `json:"turno_tipo,omitempty"`
	FaseTerminaEn  *time.Time `json:"fase_termina_en,omitempty"`
	PuntuacionA    *float64   `json:"puntuacion_a,omitempty"`
	PuntuacionB    *float64   `json:"puntuacion_b,omitempty"`
	GanadorEquipo  *string    `json:"ganador_equipo,omitempty"`
}

// Motor orquesta la ejecución en vivo de las fases de un debate: presentación, moción,
// votaciones y rondas de argumentos/réplicas/contrarréplicas, con sus tiempos y turnos.
type Motor struct {
	db      *sql.DB
	hub     *live.Hub // hub de streaming: se reutiliza para difundir los cambios de fase
	mu      sync.Mutex
	activos map[int64]chan struct{} // debateID -> canal de cancelación
}

// Nuevo crea un motor de fases listo para arrancar debates
func Nuevo(db *sql.DB, hubStreaming *live.Hub) *Motor {
	return &Motor{db: db, hub: hubStreaming, activos: make(map[int64]chan struct{})}
}

// Iniciar arranca, en una goroutine aparte, la ejecución completa de fases de un debate que
// acaba de pasar a estado "transmitiendo". No hace nada si ya había una ejecución en curso para
// ese debate (evita arrancarlo dos veces si el handler se llamara por error más de una vez).
func (m *Motor) Iniciar(debateID int64) {

	m.mu.Lock()

	if _, yaActivo := m.activos[debateID]; yaActivo {
		m.mu.Unlock()
		return
	}

	cancelar := make(chan struct{})
	m.activos[debateID] = cancelar

	m.mu.Unlock()

	go m.ejecutar(debateID, cancelar)
}

// Cancelar detiene la ejecución de fases de un debate (usado si alguien lo finaliza a mano antes
// de que termine solo)
func (m *Motor) Cancelar(debateID int64) {

	m.mu.Lock()
	defer m.mu.Unlock()

	if cancelar, existe := m.activos[debateID]; existe {
		close(cancelar)
		delete(m.activos, debateID)
	}
}

func (m *Motor) terminar(debateID int64) {
	m.mu.Lock()
	delete(m.activos, debateID)
	m.mu.Unlock()
}

// esperar bloquea por la duración indicada, o hasta que se cancele el debate. Retorna false si
// se canceló, para que el llamador corte la ejecución del resto de fases.
func esperar(d time.Duration, cancelar <-chan struct{}) bool {

	temporizador := time.NewTimer(d)
	defer temporizador.Stop()

	select {
	case <-temporizador.C:
		return true
	case <-cancelar:
		return false
	}
}

// ejecutar corre de principio a fin toda la secuencia de un debate. Se detiene apenas el canal
// de cancelación se cierra (debate finalizado a mano) o si hay un error irrecuperable.
func (m *Motor) ejecutar(debateID int64, cancelar <-chan struct{}) {

	defer m.terminar(debateID)

	debate, err := database.ObtenerDebatePorID(m.db, debateID)

	if err != nil {
		logs.Error("Motor: no se pudo cargar el debate " + strconv.FormatInt(debateID, 10) + ": " + err.Error())
		return
	}

	participantes, err := database.ListarParticipantes(m.db, debateID)

	if err != nil {
		logs.Error("Motor: no se pudieron cargar los participantes: " + err.Error())
		return
	}

	jurado, err := database.ObtenerJurado(m.db, debateID)

	if err != nil {
		logs.Error("Motor: no se pudo cargar el jurado: " + err.Error())
		return
	}

	equipoA := filtrarEquipo(participantes, database.EquipoA)
	equipoB := filtrarEquipo(participantes, database.EquipoB)

	// ---------- Fase: presentación (15s cada uno) — jurado, luego equipo A, luego equipo B ----------
	for _, miembro := range jurado {
		if !m.turno(debateID, database.FasePresentacion, 0, ref(miembro.UsuarioID), nil, strp(database.TurnoPresentacion), duracionPresentacion, cancelar) {
			return
		}
	}

	for _, p := range equipoA {
		if !m.turno(debateID, database.FasePresentacion, 0, ref(p.UsuarioID), strp(database.EquipoA), strp(database.TurnoPresentacion), duracionPresentacion, cancelar) {
			return
		}
	}

	for _, p := range equipoB {
		if !m.turno(debateID, database.FasePresentacion, 0, ref(p.UsuarioID), strp(database.EquipoB), strp(database.TurnoPresentacion), duracionPresentacion, cancelar) {
			return
		}
	}

	// ---------- Fase: moción — cada equipo plantea su postura ----------
	if vocero := primerOrador(equipoA); vocero != nil {
		if !m.turno(debateID, database.FaseMocion, 0, ref(vocero.UsuarioID), strp(database.EquipoA), strp(database.TurnoMocion), duracionMocion, cancelar) {
			return
		}
	}

	if vocero := primerOrador(equipoB); vocero != nil {
		if !m.turno(debateID, database.FaseMocion, 0, ref(vocero.UsuarioID), strp(database.EquipoB), strp(database.TurnoMocion), duracionMocion, cancelar) {
			return
		}
	}

	// ---------- Primera votación pública ----------
	if !m.turno(debateID, database.FaseVotacion1, 0, nil, nil, nil, duracionVotacion1, cancelar) {
		return
	}

	// ---------- Rondas de debate: en cada ronda, A lidera y luego B lidera ----------
	for ronda := 1; ronda <= debate.Rondas; ronda++ {

		if !m.cicloRonda(debateID, ronda, database.EquipoA, database.EquipoB, equipoA, equipoB, cancelar) {
			return
		}

		if !m.cicloRonda(debateID, ronda, database.EquipoB, database.EquipoA, equipoB, equipoA, cancelar) {
			return
		}
	}

	// ---------- Segunda votación pública + puntuación del jurado (en simultáneo) ----------
	if !m.turno(debateID, database.FaseVotacion2, debate.Rondas, nil, nil, nil, duracionVotacion2, cancelar) {
		return
	}

	m.finalizar(debateID)
}

// cicloRonda ejecuta el bloque de argumento principal, réplica y contrarréplica de un lado del
// debate dentro de una ronda: equipoLidera plantea el argumento y cierra con la contrarréplica;
// equipoResponde interviene solo con la réplica en el medio.
func (m *Motor) cicloRonda(debateID int64, ronda int, equipoLidera string, equipoResponde string, participantesLidera []database.Participante, participantesResponde []database.Participante, cancelar <-chan struct{}) bool {

	oradorLidera := elegirOrador(participantesLidera, ronda)
	oradorResponde := elegirOrador(participantesResponde, ronda)

	if !esperar(pausaAntesDeArgumento, cancelar) {
		return false
	}

	if oradorLidera != nil {
		if !m.turno(debateID, database.FaseDebate, ronda, ref(oradorLidera.UsuarioID), strp(equipoLidera), strp(database.TurnoArgumento), duracionArgumento, cancelar) {
			return false
		}
	}

	if !esperar(pausaAntesDeReplica, cancelar) {
		return false
	}

	if oradorResponde != nil {
		if !m.turno(debateID, database.FaseDebate, ronda, ref(oradorResponde.UsuarioID), strp(equipoResponde), strp(database.TurnoReplica), duracionReplica, cancelar) {
			return false
		}
	}

	if !esperar(pausaAntesDeContrarreplica, cancelar) {
		return false
	}

	if oradorLidera != nil {
		if !m.turno(debateID, database.FaseDebate, ronda, ref(oradorLidera.UsuarioID), strp(equipoLidera), strp(database.TurnoContrarreplica), duracionContrarreplica, cancelar) {
			return false
		}
	}

	return true
}

// turno guarda en la base de datos el nuevo turno/fase, lo difunde por el hub de streaming, y
// espera la duración indicada (o hasta cancelación). Retorna false si se canceló.
func (m *Motor) turno(debateID int64, fase string, ronda int, turnoUsuarioID *int64, turnoEquipo *string, turnoTipo *string, duracion time.Duration, cancelar <-chan struct{}) bool {

	terminaEn := time.Now().Add(duracion)

	err := database.ActualizarFase(m.db, debateID, fase, ronda, turnoUsuarioID, turnoEquipo, turnoTipo, &terminaEn)

	if err != nil {
		logs.Error("Motor: no se pudo guardar la fase: " + err.Error())
	}

	m.difundir(mensajeFase{
		Tipo: "fase", Fase: fase, RondaActual: ronda,
		TurnoUsuarioID: turnoUsuarioID, TurnoEquipo: turnoEquipo, TurnoTipo: turnoTipo,
		FaseTerminaEn: &terminaEn,
	}, debateID)

	return esperar(duracion, cancelar)
}

// finalizar calcula el resultado ponderado (jurado 70/100 + público 30/100 de la segunda
// votación) de cada equipo, guarda el ganador y lo difunde
func (m *Motor) finalizar(debateID int64) {

	promedioJuradoA, promedioJuradoB, err := database.PromedioJuradoPorEquipo(m.db, debateID)

	if err != nil {
		logs.Error("Motor: no se pudo calcular el promedio del jurado: " + err.Error())
	}

	resultado, err := database.ObtenerResultados(m.db, debateID, 2)

	if err != nil {
		logs.Error("Motor: no se pudieron obtener los resultados públicos: " + err.Error())
	}

	// jurado: promedio (0-10) escalado a 70 puntos · público: porcentaje de votos escalado a 30 puntos
	puntuacionA := (promedioJuradoA/10)*70 + (resultado.PorcentajeA/100)*30
	puntuacionB := (promedioJuradoB/10)*70 + (resultado.PorcentajeB/100)*30

	ganador := database.EquipoA

	if puntuacionB > puntuacionA {
		ganador = database.EquipoB
	}

	err = database.FinalizarConResultado(m.db, debateID, puntuacionA, puntuacionB, ganador)

	if err != nil {
		logs.Error("Motor: no se pudo finalizar el debate con resultado: " + err.Error())
	}

	m.difundir(mensajeFase{
		Tipo: "fase", Fase: database.FaseFinalizado,
		PuntuacionA: &puntuacionA, PuntuacionB: &puntuacionB, GanadorEquipo: &ganador,
	}, debateID)

	logs.Info("Debate " + strconv.FormatInt(debateID, 10) + " finalizado. Ganador: equipo " + ganador)
}

func (m *Motor) difundir(msg mensajeFase, debateID int64) {

	datos, err := json.Marshal(msg)

	if err != nil {
		logs.Error("Motor: no se pudo serializar el mensaje de fase: " + err.Error())
		return
	}

	m.hub.Difundir(debateID, string(datos), nil)
}

// ---------- Auxiliares ----------

func filtrarEquipo(participantes []database.Participante, equipo string) []database.Participante {

	filtrados := make([]database.Participante, 0)

	for _, p := range participantes {
		if p.Equipo == equipo {
			filtrados = append(filtrados, p)
		}
	}

	return filtrados
}

func primerOrador(participantes []database.Participante) *database.Participante {

	if len(participantes) == 0 {
		return nil
	}

	return &participantes[0]
}

// elegirOrador rota quién habla por equipo entre rondas (ronda 1 = índice 0, ronda 2 = índice 1, …)
// para repartir el turno de forma pareja si el equipo tiene más de un integrante.
func elegirOrador(participantes []database.Participante, ronda int) *database.Participante {

	if len(participantes) == 0 {
		return nil
	}

	indice := (ronda - 1) % len(participantes)

	return &participantes[indice]
}

func ref(v int64) *int64    { return &v }
func strp(v string) *string { return &v }
