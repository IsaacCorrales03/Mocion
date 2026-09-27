package handlers

// Importaciones de librerías necesarias

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/IsaacCorrales03/Mocion/backend/internal/auth"
)

// loginPeticion representa el cuerpo JSON esperado para iniciar sesión
type loginPeticion struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginRespuesta representa el cuerpo JSON de una respuesta exitosa de login
type loginRespuesta struct {
	Token  string `json:"token"`
	ID     int64  `json:"id"`
	Nombre string `json:"nombre"`
	Email  string `json:"email"`
}

// NuevoLoginHandler crea el handler de login inyectando la conexión a la base de datos
func NuevoLoginHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			responderError(w, http.StatusMethodNotAllowed, "método no permitido")
			return
		}

		var peticion loginPeticion

		err := json.NewDecoder(r.Body).Decode(&peticion)

		if err != nil {
			responderError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
			return
		}

		token, usuario, err := auth.IniciarSesion(db, peticion.Email, peticion.Password)

		if err == auth.ErrCredencialesInvalidas {
			responderError(w, http.StatusUnauthorized, err.Error())
			return
		}

		if err != nil {
			responderError(w, http.StatusInternalServerError, "error interno al iniciar sesión")
			return
		}

		respuesta := loginRespuesta{
			Token:  token,
			ID:     usuario.ID,
			Nombre: usuario.Nombre,
			Email:  usuario.Email,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(respuesta)
	}
}
