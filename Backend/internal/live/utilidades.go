package live

// Importaciones de librerías necesarias

import "strconv"

// itoa convierte un ID de tipo int64 a texto, usado únicamente para mensajes de log
func itoa(id int64) string {
	return strconv.FormatInt(id, 10)
}
