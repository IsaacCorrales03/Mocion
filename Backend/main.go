package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/IsaacCorrales03/Mocion/backend/database"
	"github.com/IsaacCorrales03/Mocion/backend/internal/handlers"
	"github.com/IsaacCorrales03/Mocion/backend/internal/live"
	"github.com/IsaacCorrales03/Mocion/backend/internal/logs"
	"github.com/IsaacCorrales03/Mocion/backend/internal/middleware"
	"github.com/IsaacCorrales03/Mocion/backend/internal/motor"
	"github.com/IsaacCorrales03/Mocion/backend/internal/presencia"
	"github.com/joho/godotenv"
)

func main() {

	logs.Info("Iniciando servidor")

	err := godotenv.Load()

	if err != nil {
		// En Render (y otros PaaS) no hay archivo .env: las variables las
		// inyecta la plataforma directamente. Por eso esto solo se avisa,
		// nunca corta el arranque del servidor.
		logs.Info("No se encontró archivo .env, se usan las variables de entorno del sistema")
	}

	URLDeLaBaseDeDatos := os.Getenv("URL_BASE_DE_DATOS")

	db, err := database.Conectar(URLDeLaBaseDeDatos)

	if err != nil {
		return
	}

	defer db.Close()

	err = prepararTablas(db)

	if err != nil {
		return
	}

	// Hub para el chat en vivo y hub separado para la señalización de streaming,
	// cada uno administra sus propias salas por ID de debate
	hubChat := live.NuevoHub()
	hubStreaming := live.NuevoHub()

	// Motor de fases: orquesta presentación, moción, votaciones y rondas de debate en vivo,
	// reutilizando el hub de streaming para avisar a todos los conectados de cada cambio de turno
	motorDeFases := motor.Nuevo(db, hubStreaming)

	// Registro de presencia: quién tiene la página del debate abierta en este momento, para
	// exigir que todos los asignados estén realmente conectados antes de poder arrancar
	registroPresencia := presencia.Nuevo()

	http.HandleFunc("/", index)

	// Autenticación
	http.HandleFunc("/login", handlers.NuevoLoginHandler(db))
	http.HandleFunc("/register", handlers.NuevoRegisterHandler(db))

	// Usuarios
	http.HandleFunc("GET /usuarios/buscar", handlers.NuevoBuscarUsuariosHandler(db))

	// Debates
	http.HandleFunc("POST /debates", handlers.NuevoCrearDebateHandler(db))
	http.HandleFunc("GET /debates/en-vivo", handlers.NuevoExplorarDebatesHandler(db))
	http.HandleFunc("GET /debates/programados", handlers.NuevoDebatesProgramadosHandler(db))
	http.HandleFunc("GET /debates/{id}", handlers.NuevoObtenerDebateHandler(db))
	http.HandleFunc("POST /debates/{id}/iniciar", handlers.NuevoIniciarDebateHandler(db, motorDeFases, registroPresencia))
	http.HandleFunc("POST /debates/{id}/finalizar", handlers.NuevoFinalizarDebateHandler(db, motorDeFases))

	// Presencia: quién tiene la página abierta ahora mismo
	http.HandleFunc("/debates/{id}/presencia", handlers.NuevoPresenciaHandler(registroPresencia))
	http.HandleFunc("GET /debates/{id}/conectados", handlers.NuevoConectadosHandler(registroPresencia))

	// Participantes (equipos A/B)
	http.HandleFunc("POST /debates/{id}/participantes", handlers.NuevoAsignarParticipanteHandler(db))
	http.HandleFunc("GET /debates/{id}/participantes", handlers.NuevoListarParticipantesHandler(db))

	// Votaciones (1ª y 2ª ronda, por equipo)
	http.HandleFunc("POST /debates/{id}/votar", handlers.NuevoVotarHandler(db))
	http.HandleFunc("GET /debates/{id}/resultados", handlers.NuevoResultadosHandler(db))

	// Jurado y puntuaciones
	http.HandleFunc("POST /debates/{id}/jurado", handlers.NuevoAsignarJuradoHandler(db))
	http.HandleFunc("GET /debates/{id}/jurado", handlers.NuevoObtenerJuradoHandler(db))
	http.HandleFunc("POST /debates/{id}/puntuacion", handlers.NuevoPuntuarHandler(db))
	http.HandleFunc("GET /debates/{id}/puntuaciones", handlers.NuevoObtenerPuntuacionesHandler(db))

	// Tiempo real: chat en vivo y señalización del streaming + avisos de fase (WebSocket)
	http.HandleFunc("/debates/{id}/chat", handlers.NuevoChatHandler(hubChat))
	http.HandleFunc("/debates/{id}/stream", handlers.NuevoStreamingHandler(hubStreaming))

	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "8080" // fallback para correrlo en local
	}

	logs.Info("Servidor iniciado en el puerto " + puerto)

	err = http.ListenAndServe(":"+puerto, middleware.CORS(http.DefaultServeMux))

	if err != nil {
		logs.Critical("El servidor se detuvo: " + err.Error())
	}
}

// prepararTablas crea todas las tablas necesarias para el MVP, en orden por sus dependencias (foreign keys)
func prepararTablas(db *sql.DB) error {

	err := database.CrearTablaUsuarios(db)

	if err != nil {
		logs.Critical("No se pudo preparar la tabla de usuarios: " + err.Error())
		return err
	}

	err = database.CrearTablaDebates(db)

	if err != nil {
		logs.Critical("No se pudo preparar la tabla de debates: " + err.Error())
		return err
	}

	err = database.MigrarTablaDebates(db)

	if err != nil {
		logs.Critical("No se pudo migrar la tabla de debates: " + err.Error())
		return err
	}

	err = database.CrearTablaParticipantes(db)

	if err != nil {
		logs.Critical("No se pudo preparar la tabla de participantes: " + err.Error())
		return err
	}

	err = database.CrearTablaVotos(db)

	if err != nil {
		logs.Critical("No se pudo preparar la tabla de votos: " + err.Error())
		return err
	}

	err = database.MigrarTablaVotos(db)

	if err != nil {
		logs.Critical("No se pudo migrar la tabla de votos: " + err.Error())
		return err
	}

	err = database.CrearTablaJurado(db)

	if err != nil {
		logs.Critical("No se pudo preparar la tabla de jurado: " + err.Error())
		return err
	}

	err = database.CrearTablaPuntuaciones(db)

	if err != nil {
		logs.Critical("No se pudo preparar la tabla de puntuaciones: " + err.Error())
		return err
	}

	return nil
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "iniciado")
}

