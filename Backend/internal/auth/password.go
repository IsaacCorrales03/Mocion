package auth

// Importaciones de librerías necesarias

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// generarSalt crea una sal aleatoria de 16 bytes en formato hexadecimal
func generarSalt() (string, error) {

	bytesAleatorios := make([]byte, 16)

	_, err := rand.Read(bytesAleatorios)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytesAleatorios), nil
}

// HashPassword genera el hash de una contraseña junto con una sal nueva
// Nota: se usa sha256 + sal por no contar con acceso al proxy de módulos de Go
// para descargar golang.org/x/crypto/bcrypt. Si se habilita esa dependencia,
// se recomienda migrar a bcrypt.
func HashPassword(password string) (hash string, salt string, err error) {

	salt, err = generarSalt()

	if err != nil {
		return "", "", err
	}

	hash = calcularHash(password, salt)

	return hash, salt, nil
}

// VerificarPassword comprueba si una contraseña coincide con el hash y la sal guardados
func VerificarPassword(password string, hash string, salt string) bool {

	hashCalculado := calcularHash(password, salt)

	return hashCalculado == hash
}

// calcularHash combina la contraseña y la sal y calcula el hash sha256
func calcularHash(password string, salt string) string {

	datos := sha256.Sum256([]byte(password + salt))

	return hex.EncodeToString(datos[:])
}
