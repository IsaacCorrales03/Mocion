package logs

// Importaciones de librerías necesarias
import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Función que nos dará la fecha para el archivo log
func ObtenerFecha() string {
	return time.Now().Format("02-01-2006")
}

// Función que nos va a dar la fecha y la hora para el logger
func ObtenerHoraYFecha() string {
	// Obtenemos el día, mes y año así como hora, minuto y segundo
	var fechaConHora string = time.Now().Format("02-01-2006 15:04:05")
	// Concatenamos y añadimos el "[]" para formato bonito
	return "[" + fechaConHora + "]"
}

// Función que imprimirá los logs en el archivo fecha.log
func EscribirEnArchivo(mensaje string) {
	// Primero, construimos la ruta del archivo log:
	var fecha string = ObtenerFecha()
	var ruta string = filepath.Join("Logs", fecha+".log")

	// comprobamos si el archivo existe:
	_, err := os.Stat(ruta)
	if os.IsNotExist(err) {
		// Creamos el archivo:
		archivo, err := os.Create(ruta)
		if err != nil {
			return
		}
		archivo.Close()
	}
	// abrimos el archivo:
	archivo, err := os.OpenFile(ruta, os.O_APPEND|os.O_WRONLY, 0644)
	// si hubo un err al abrir, terminamos la función
	if err != nil {
		return
	}
	// escribimos el contenido
	archivo.WriteString(mensaje + "\n")
	// cerramos el archivo
	archivo.Close()
}

func Info(mensaje string) {
	// Función que logea información
	var texto string = ObtenerHoraYFecha() + " [INFO] " + mensaje
	fmt.Println(texto)
	EscribirEnArchivo(texto)
}

func Error(mensaje string) {
	// Función que logea errores
	var texto string = ObtenerHoraYFecha() + " [ERROR] " + mensaje
	fmt.Println(texto)
	EscribirEnArchivo(texto)
}

func Critical(mensaje string) {
	// Función que logea errores criticos y de urgencia
	var texto string = ObtenerHoraYFecha() + " [CRITICAL] " + mensaje
	fmt.Println(texto)
	EscribirEnArchivo(texto)
}
