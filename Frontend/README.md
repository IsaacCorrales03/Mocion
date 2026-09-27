# Moción — Frontend

Frontend estático (HTML/CSS/JS sin build) para el backend Go de `mocion-backend`, con la
identidad visual definida en el brief: editorial dark UI, bento asimétrico, acento único.

## Estructura
```
index.html        Debates en directo + abiertos
login.html          Iniciar sesión / crear cuenta
crear.html          Programar un nuevo debate (título, rondas, tamaño de equipo)
debate.html         Gestión de equipos/jurado, vista en vivo por fases, y resultado final
css/style.css       Sistema visual compartido (color, tipografía, componentes)
js/api.js           Cliente de la API (fetch + WebSocket)
js/ui.js            Utilidades de interfaz (reloj, sesión, formatos)
js/pages/*.js        Lógica de cada página
```

## Uso
Servir la carpeta con cualquier servidor estático (no puede abrirse con `file://` por los
módulos y `fetch`). Incluye un `serve.json` que desactiva las "clean URLs" — necesario
porque, si `npx serve` reescribe `debate.html` a `/debate`, esa ruta sin extensión termina
resolviendo a `index.html` en vez de al archivo correcto. Corré:
```
npx serve .
```
**No uses el flag `-s`** (modo single-page app): fuerza a `serve` a devolver siempre
`index.html` para cualquier ruta que no sea un archivo exacto, lo cual rompe la navegación
a `debate.html`, `login.html` y `crear.html`.

Por defecto la API apunta al mismo host desde el que se abrió la página, puerto `8080`
(así funciona tanto en `localhost` como entrando por la IP de tu red). Para apuntar a otro
host, definir antes de cargar `api.js`:
```html
<script>window.MOCION_API_URL = "https://tu-servidor.com";</script>
```
**Cámara/micrófono solo funcionan en `localhost` o HTTPS** — es una restricción del
navegador, no de esta app. Si entrás por una IP de red (`http://192.168.x.x:3000`), vas a
poder mirar la transmisión pero no vas a poder transmitir tu propia cámara desde ahí.

## La lógica del debate (estado + fases)

Un debate pasa por tres **estados**: `abierto` → `transmitiendo` → `finalizado`. Mientras
está `transmitiendo`, el motor de fases del backend (`internal/motor`) va avanzando solo,
con temporizadores reales, por estas **fases**, en este orden:

1. **`presentacion`** — 15s por persona: jurado (si hay), luego equipo A, luego equipo B.
2. **`mocion`** — cada equipo plantea su postura (60s cada uno — *asumido*, no se
   especificó un tiempo exacto).
3. **`votacion_1`** — el público vota A o B durante 30s.
4. **`debate`** — se repite `rondas` veces (1 a 5, elegido al crear el debate). En cada
   ronda, primero lidera el equipo A y después el equipo B, con esta secuencia por cada
   lado: argumento principal (90s) → réplica del otro equipo (60s) → contrarréplica (30s).
   Entre cada turno hay una pausa (60s antes de argumento/réplica, 30s antes de
   contrarréplica) durante la cual no hay nadie transmitiendo — el próximo orador puede
   pulsar "TOMAR LA PALABRA" apenas el turno se lo asigna.
5. **`votacion_2`** — el público vota de nuevo (90s — *asumido*) mientras, en simultáneo,
   el jurado califica de 1 a 10 a cada participante.
6. **`finalizado`** — se calcula el resultado y se cierra el debate.

**Fórmula del resultado** (tal como se pidió): por equipo, `(promedio del jurado / 10) × 70
+ (porcentaje de la 2ª votación / 100) × 30`, sobre 100 puntos. Gana el equipo con más
puntos.

Todo esto se transmite en vivo por el mismo socket de streaming (`/debates/{id}/stream`)
como mensajes `{"tipo": "fase", ...}`, así que cualquiera conectado ve la cuenta regresiva,
de quién es el turno, y cuándo abren las votaciones, sin recargar la página.

## Aviso sobre el contrato de la API

- **`POST /login` ya devuelve `id`.** (arreglado en una vuelta anterior)
- El chat y la señalización de streaming siguen sin un formato de mensaje definido por el
  backend; uso las mismas convenciones (`{"usuario","texto"}` y
  `{"tipo":"oferta"|"respuesta"|"candidato"|"fase", ...}`) documentadas antes.
- **Streaming: 1 emisor : 1 espectador por vez** (limitación ya conocida del hub, sin
  enrutamiento por destinatario). Con el nuevo flujo esto es menos grave que antes: en la
  práctica solo habla una persona a la vez (el motor asigna un único `turno_usuario_id`),
  así que el problema real sería tener *varios espectadores* mirando al mismo tiempo — ahí
  sigue aplicando la misma limitación: solo el primero que responda completa la conexión.
- **La calificación del jurado no está bloqueada por fase del lado del servidor.** La
  interfaz solo muestra el panel de puntuación durante `votacion_2` y solo a quienes son
  jurado, pero el endpoint `POST /debates/{id}/puntuacion` no comprueba en qué fase está el
  debate — alguien con el `usuario_id` correcto podría llamarlo en cualquier momento. Si
  esto importa (por ejemplo, si el frontend termina siendo público y no solo un cliente de
  confianza), conviene agregar esa validación en el handler.
- **La cuenta de "participantes/jurado por equipo" no depende de si esas personas ya están
  conectadas** — un organizador puede marcar el debate como completo y arrancarlo aunque
  alguien del equipo no haya abierto la página todavía; en ese caso, cuando le toque su
  turno, simplemente no va a haber nadie para pulsar "TOMAR LA PALABRA" y ese turno pasa
  igual, en silencio, cuando se acaba el tiempo.
- **Migración de base de datos:** si ya tenías una base de datos de una versión anterior
  (con `votos.opcion` en `'a_favor'/'en_contra'` o `jurado.veredicto`), la migración nueva
  (`MigrarTablaVotos`) NO convierte esos datos viejos al esquema A/B — solo ajusta la
  estructura de las columnas. Si tenés votos o veredictos de pruebas anteriores que no te
  importan, lo más simple es limpiar esas tablas (`TRUNCATE votos, jurado;`) después de
  actualizar.
