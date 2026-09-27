# Mocion

Mocion es una plataforma web que busca ofrecer a los jóvenes un espacio para debatir.

La idea principal es realizar debates en vivo con público, chat, votaciones y jurado.

## Tecnologías

* Go
* React
* HTML
* CSS
* JavaScript

## MVP

* Login y registro
* Video de prueba
* Crear debates
* Streaming en vivo
* Chat en vivo
* Votaciones
* Jurado
* Explorar debates en vivo



Documentacion:
backend:
dentro tiene las carpetas:
cmd
database
internal

# main.go:
Contiene el servidor principal

# database: 
contiene la base de datos así como sus operaciones (obtener usuarios, etc etc)

# Internal: 
tiene la lógica interna del servidor, con las carpetas:
auth
handlers

# Internal/auth
esta carpeta contiene la lógica de autenticación de usuario
login.go:
se encarga de iniciar sesión, crear una sesion, y manejar tokens
register.go:
se encarga de crear, validar y añadir un usuario a la base de datos

# Internal/handlers
cuando se realiza una petición, esta es procesada por un handler, esta 
carpeta contiene todos los handlers

# Internal/logs
carpeta que tiene el logger del servidor, logs que se guardan en la carpeta Logs