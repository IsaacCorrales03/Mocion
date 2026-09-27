package handlers

// Importaciones de librerías necesarias

import (
	"encoding/json"
	"net/http"
)

// errorRespuesta representa el cuerpo JSON de una respuesta de error
type errorRespuesta struct {
	Error string `json:"error"`
}

// responderError escribe una respuesta de error en formato JSON con el código de estado indicado
func responderError(w http.ResponseWriter, codigo int, mensaje string) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(codigo)
	json.NewEncoder(w).Encode(errorRespuesta{Error: mensaje})
}

// responderJSON escribe cualquier estructura como respuesta JSON con el código de estado indicado
func responderJSON(w http.ResponseWriter, codigo int, cuerpo interface{}) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(codigo)
	json.NewEncoder(w).Encode(cuerpo)
}
