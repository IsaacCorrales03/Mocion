package database

// Importaciones de librerías necesarias

import (
	"database/sql"
	"github.com/IsaacCorrales03/Mocion/backend/internal/logs"
	_ "github.com/lib/pq"
)

// Creamos la función para crear una conexión

func Conectar(URLDeLaBaseDeDatos string) (*sql.DB, error) {

	// Informamos que estamos intentando conectar

	logs.Info("Conectando con la DB")

	// Abrimos el objeto sql.DB

	BaseDeDatos, err := sql.Open("postgres", URLDeLaBaseDeDatos)

	// Si el err no es nil, algo salió mal

	if err != nil {

		logs.Error("Error al abrir la conexión con la DB: " + err.Error())

		return nil, err
	}

	// Comprobamos la conexión

	err = BaseDeDatos.Ping()

	// Si nos da un err

	if err != nil {

		logs.Error("Error al conectar con la DB: " + err.Error())

		return nil, err
	}

	// La conexión es válida

	logs.Info("Conectado a la DB")

	// Devolvemos la conexión y nil como err

	return BaseDeDatos, nil
}