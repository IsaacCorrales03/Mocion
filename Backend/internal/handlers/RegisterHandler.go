package handlers

// Importaciones de librerías necesarias

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/IsaacCorrales03/Mocion/backend/database"
	"github.com/IsaacCorrales03/Mocion/backend/internal/auth"
)

// registroPeticion representa el cuerpo JSON esperado para registrar un usuario
type registroPeticion struct {
	Nombre   string `json:"nombre"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// registroRespuesta representa el cuerpo JSON de una respuesta exitosa de registro
type registroRespuesta struct {
	ID     int64  `json:"id"`
	Nombre string `json:"nombre"`
	Email  string `json:"email"`
}

// NuevoRegisterHandler crea el handler de registro inyectando la conexión a la base de datos
func NuevoRegisterHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			responderError(w, http.StatusMethodNotAllowed, "método no permitido")
			return
		}

		var peticion registroPeticion

		err := json.NewDecoder(r.Body).Decode(&peticion)

		if err != nil {
			responderError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}

		usuario, err := auth.RegistrarUsuario(db, peticion.Nombre, peticion.Email, peticion.Password)

		if err == auth.ErrDatosInvalidos || err == auth.ErrPasswordCorta {
			responderError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err == database.ErrCorreoYaRegistrado {
			responderError(w, http.StatusConflict, err.Error())
			return
		}

		if err != nil {
			responderError(w, http.StatusInternalServerError, "error interno al registrar el usuario")
			return
		}

		respuesta := registroRespuesta{
			ID:     usuario.ID,
			Nombre: usuario.Nombre,
			Email:  usuario.Email,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(respuesta)
	}
}
