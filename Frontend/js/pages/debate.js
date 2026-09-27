iniciarReloj("clock");
iniciarReloj("clockDebate");
pintarSesionEnNav();
const sesion = requerirSesion();

const debateId = location.hash.slice(1);
if (!debateId) location.href = "index.html";

let debateActual = null;
let participantesCache = [];
let juradoCache = [];
let esJurado = false;
let miTurno = false;

const NOMBRES_FASE = {
  "presentacion": "Presentación",
  "mocion": "Planteamiento de la moción",
  "votacion_1": "Primera votación del público",
  "debate": "Debate",
  "votacion_2": "Segunda votación y puntuación del jurado",
  "finalizado": "Debate finalizado",
};
const NOMBRES_TURNO = {
  "presentacion": "se presenta",
  "mocion": "plantea la moción",
  "argumento": "argumento principal",
  "replica": "réplica",
  "contrarreplica": "contrarréplica",
};

/* =====================================================================
   Carga inicial: decide qué panel mostrar según el estado del debate
   ===================================================================== */
async function cargarDebate() {
  try {
    debateActual = await API.obtenerDebate(debateId);
  } catch (e) {
    document.getElementById("cargando").textContent = `No se pudo cargar el debate (${e.message}).`;
    return;
  }

  document.getElementById("cargando").style.display = "none";
  document.getElementById("contenido").style.display = "";
  document.getElementById("debateId").textContent = String(debateActual.id).padStart(4, "0");
  document.getElementById("titulo").textContent = debateActual.titulo;
  document.getElementById("descripcion").textContent = debateActual.descripcion || "Sin descripción adicional.";

  try {
    participantesCache = await API.listarParticipantes(debateId);
    juradoCache = await API.obtenerJurado(debateId);
    esJurado = juradoCache.some(j => j.usuario_id === sesion.id);
  } catch { /* silencioso */ }

  if (debateActual.estado === "abierto") {
    mostrarPanelGestion();
  } else if (debateActual.estado === "transmitiendo") {
    mostrarPanelVivo();
  } else {
    mostrarPanelFinal();
  }
}

function marcarEstado(texto) {
  document.getElementById("estadoMark").innerHTML =
    debateActual.estado === "transmitiendo"
      ? '<span class="live-dot"></span> EN DIRECTO'
      : `<span style="color:var(--ink-faint)">${texto}</span>`;
}

/* =====================================================================
   Panel 1 — ABIERTO: armar equipos y jurado antes de arrancar
   ===================================================================== */
function mostrarPanelGestion() {
  marcarEstado("ABIERTO");
  document.getElementById("panelGestion").style.display = "";
  pintarListasGestion();
  conectarBuscadores();
}

function pintarListasGestion() {
  const equipoA = participantesCache.filter(p => p.equipo === "a");
  const equipoB = participantesCache.filter(p => p.equipo === "b");
  const req = debateActual.participantes_por_equipo;

  document.getElementById("conteoA").textContent = `${equipoA.length} / ${req}`;
  document.getElementById("conteoB").textContent = `${equipoB.length} / ${req}`;
  document.getElementById("conteoJurado").textContent = `${juradoCache.length} asignados`;

  pintarPersonas("listaEquipoA", equipoA);
  pintarPersonas("listaEquipoB", equipoB);
  pintarPersonas("listaJuradoGestion", juradoCache);
}

function pintarPersonas(contenedorId, lista) {
  const cont = document.getElementById(contenedorId);
  cont.innerHTML = lista.length ? "" : '<div style="font-family:var(--mono);font-size:11px;color:var(--ink-faint);">Nadie asignado todavía.</div>';
  lista.forEach(p => {
    const row = document.createElement("div");
    row.className = "persona-row";
    row.innerHTML = `<span>${p.nombre}</span>`;
    cont.appendChild(row);
  });
}

function conectarBuscadores() {
  document.querySelectorAll(".buscador input").forEach(input => {
    let temporizador;
    input.addEventListener("input", () => {
      clearTimeout(temporizador);
      temporizador = setTimeout(() => buscarYMostrar(input), 300);
    });
  });
}

async function buscarYMostrar(input) {
  const destino = input.dataset.destino;
  const cont = input.parentElement.querySelector(".resultados-busqueda");
  const texto = input.value.trim();

  if (texto.length < 2) { cont.innerHTML = ""; return; }

  let usuarios;
  try { usuarios = await API.buscarUsuarios(texto); } catch { return; }

  cont.innerHTML = "";
  usuarios.forEach(u => {
    const row = document.createElement("div");
    row.className = "resultado-usuario";
    row.innerHTML = `<span>${u.nombre}<br><span style="color:var(--ink-faint)">${u.email}</span></span><button>AGREGAR</button>`;
    row.querySelector("button").addEventListener("click", async () => {
      try {
        if (destino === "jurado") await API.asignarJurado(debateId, u.id);
        else await API.asignarParticipante(debateId, u.id, destino);
        input.value = ""; cont.innerHTML = "";
        participantesCache = await API.listarParticipantes(debateId);
        juradoCache = await API.obtenerJurado(debateId);
        pintarListasGestion();
      } catch (e) { alert(e.message); }
    });
    cont.appendChild(row);
  });
}

document.getElementById("btnIniciar").addEventListener("click", async () => {
  try {
    await API.iniciarDebate(debateId, false);
    location.reload();
  } catch (e) {
    if (e.message.includes("no están todos")) {
      if (confirm("No están todos los miembros todavía, ¿seguro que querés iniciarlo?")) {
        try { await API.iniciarDebate(debateId, true); location.reload(); }
        catch (e2) { alert(e2.message); }
      }
    } else {
      alert(e.message);
    }
  }
});

/* =====================================================================
   Panel 2 — TRANSMITIENDO: vista en vivo, fase a fase
   ===================================================================== */
function mostrarPanelVivo() {
  marcarEstado("EN DIRECTO");
  document.getElementById("panelVivo").style.display = "";

  const esCreador = sesion.id && sesion.id === debateActual.creador_id;
  document.getElementById("btnFinalizar").style.display = esCreador ? "" : "none";

  aplicarFase(debateActual);
  conectarChat();
  conectarStream();
  cargarResultadosVotacion(faseARonda(debateActual.fase));
}

function faseARonda(fase) {
  return fase === "votacion_2" ? 2 : 1;
}

let cuentaRegresivaInterval;
function aplicarFase(d) {
  const fase = d.fase;

  document.getElementById("faseNombre").textContent = NOMBRES_FASE[fase] || "Preparando…";
  document.getElementById("faseRonda").textContent = fase === "debate" ? `RONDA ${d.ronda_actual}` : "";

  clearInterval(cuentaRegresivaInterval);
  if (d.fase_termina_en) {
    const actualizarCuenta = () => {
      const restante = Math.max(0, Math.floor((new Date(d.fase_termina_en) - Date.now()) / 1000));
      const min = String(Math.floor(restante / 60)).padStart(2, "0");
      const seg = String(restante % 60).padStart(2, "0");
      document.getElementById("faseCuenta").textContent = `${min}:${seg}`;
    };
    actualizarCuenta();
    cuentaRegresivaInterval = setInterval(actualizarCuenta, 1000);
  } else {
    document.getElementById("faseCuenta").textContent = "--:--";
  }

  miTurno = !!(d.turno_usuario_id && sesion.id === d.turno_usuario_id);

  const equipoLbl = d.turno_equipo ? `Equipo ${d.turno_equipo.toUpperCase()}` : "Jurado";
  const tipoLbl = NOMBRES_TURNO[d.turno_tipo] || "";
  document.getElementById("turnoAviso").innerHTML = d.turno_usuario_id
    ? `Turno de <b>${equipoLbl}</b> — ${tipoLbl}`
    : "";
  document.getElementById("speakerTag").textContent = d.turno_usuario_id
    ? `${equipoLbl} — ${tipoLbl}`
    : "Transmisión no iniciada";

  document.getElementById("btnTomarPalabra").style.display = miTurno && !esEmisor ? "" : "none";

  const enVotacion = fase === "votacion_1" || fase === "votacion_2";
  document.getElementById("votacionPanel").style.display = enVotacion ? "" : "none";
  document.getElementById("votacionTitulo").textContent = fase === "votacion_2"
    ? "Segunda votación del público"
    : "Votación del público";

  const enPuntuacion = fase === "votacion_2" && esJurado;
  document.getElementById("puntuacionPanel").style.display = enPuntuacion ? "" : "none";
  if (enPuntuacion) pintarPanelPuntuacion();

  if (fase === "finalizado") {
    document.getElementById("panelVivo").style.display = "none";
    mostrarResultado(d);
  }
}

document.getElementById("btnFinalizar").addEventListener("click", async () => {
  if (!confirm("¿Seguro que querés finalizar el debate ahora mismo?")) return;
  try { await API.finalizarDebate(debateId); location.reload(); } catch (e) { alert(e.message); }
});

/* ---------- Votación (dos rondas, misma UI) ---------- */
async function cargarResultadosVotacion(ronda) {
  try {
    const r = await API.resultados(debateId, ronda);
    document.getElementById("pctA").textContent = Math.round(r.porcentaje_a) + "%";
    document.getElementById("pctB").textContent = Math.round(r.porcentaje_b) + "%";
    document.getElementById("barraA").style.width = r.porcentaje_a + "%";
    document.getElementById("barraB").style.width = r.porcentaje_b + "%";
  } catch { /* todavía sin votos */ }
}

document.getElementById("btnVotarA").addEventListener("click", () => votar("a"));
document.getElementById("btnVotarB").addEventListener("click", () => votar("b"));
async function votar(opcion) {
  try {
    await API.votar(debateId, sesion.id, opcion);
    await cargarResultadosVotacion(faseARonda(debateActual.fase));
  } catch (e) { alert(e.message); }
}

/* ---------- Puntuación del jurado ---------- */
function pintarPanelPuntuacion() {
  const cont = document.getElementById("listaPuntuar");
  cont.innerHTML = "";
  participantesCache.forEach(p => {
    const row = document.createElement("div");
    row.className = "puntuar-row";
    const opciones = Array.from({ length: 10 }, (_, i) => i + 1)
      .map(n => `<option value="${n}">${n}</option>`).join("");
    row.innerHTML = `<span>${p.nombre} <span style="color:var(--ink-faint);font-family:var(--mono);font-size:10px;">EQUIPO ${p.equipo.toUpperCase()}</span></span>
      <select data-usuario="${p.usuario_id}"><option value="">—</option>${opciones}</select>`;
    row.querySelector("select").addEventListener("change", async (e) => {
      const puntuacion = parseInt(e.target.value, 10);
      if (!puntuacion) return;
      try { await API.puntuar(debateId, sesion.id, p.usuario_id, puntuacion); }
      catch (err) { alert(err.message); }
    });
    cont.appendChild(row);
  });
}

/* ---------- Chat en vivo (WebSocket) ---------- */
function agregarMensaje(who, texto, animado = true) {
  const list = document.getElementById("chatList");
  const el = document.createElement("div");
  el.className = "msg" + (animado ? " msg-in" : "");
  el.innerHTML = `<span class="who">${who}</span><span class="t">${horaCorta(new Date().toISOString())}</span>${texto}`;
  list.appendChild(el);
  list.scrollTop = list.scrollHeight;
}

let chatWs;
function conectarChat() {
  chatWs = API.chatSocket(debateId);
  const estado = document.getElementById("chatEstado");
  chatWs.onopen = () => estado.textContent = "en línea";
  chatWs.onclose = () => { estado.textContent = "desconectado — reintentando…"; setTimeout(conectarChat, 2000); };
  chatWs.onerror = () => estado.textContent = "error de conexión";
  chatWs.onmessage = (ev) => {
    try {
      const data = JSON.parse(ev.data);
      agregarMensaje(data.usuario || "anónimo", data.texto || "");
    } catch {
      agregarMensaje("—", ev.data);
    }
  };
}

document.getElementById("chatSend").addEventListener("click", enviarMensaje);
document.getElementById("chatInput").addEventListener("keydown", (e) => { if (e.key === "Enter") enviarMensaje(); });
function enviarMensaje() {
  const input = document.getElementById("chatInput");
  const texto = input.value.trim();
  if (!texto || !chatWs || chatWs.readyState !== WebSocket.OPEN) return;
  chatWs.send(JSON.stringify({ usuario: sesion.nombre, texto }));
  agregarMensaje("Tú", texto);
  input.value = "";
}

/* =====================================================================
   Streaming (WebRTC) + avisos de fase — comparten el mismo socket
   ===================================================================== */
let streamWs;
let pc;
let esEmisor = false;

function conectarStream() {
  streamWs = API.streamSocket(debateId);
  streamWs.onmessage = async (ev) => {
    const data = JSON.parse(ev.data);

    if (data.tipo === "fase") {
      // El servidor difunde el avance del motor de fases por este mismo socket
      debateActual = { ...debateActual, ...data };
      if (!esEmisor) resetearVideo();
      aplicarFase(debateActual);
      return;
    }

    if (data.tipo === "oferta" && !esEmisor) {
      pc = nuevaConexion();
      await pc.setRemoteDescription(data.sdp);
      const respuesta = await pc.createAnswer();
      await pc.setLocalDescription(respuesta);
      streamWs.send(JSON.stringify({ tipo: "respuesta", sdp: respuesta }));

    } else if (data.tipo === "respuesta" && esEmisor && pc) {
      await pc.setRemoteDescription(data.sdp);

    } else if (data.tipo === "candidato" && pc) {
      try { await pc.addIceCandidate(data.candidato); } catch { /* candidato tardío, se ignora */ }
    }
  };
}

function resetearVideo() {
  const video = document.getElementById("videoRemoto");
  if (!esEmisor) video.srcObject = null;
}

function nuevaConexion() {
  const conexion = new RTCPeerConnection();
  conexion.onicecandidate = (e) => {
    if (e.candidate) streamWs.send(JSON.stringify({ tipo: "candidato", candidato: e.candidate }));
  };
  conexion.ontrack = (e) => {
    if (!esEmisor) document.getElementById("videoRemoto").srcObject = e.streams[0];
  };
  return conexion;
}

async function iniciarOferta() {
  pc = nuevaConexion();

  let local;
  try {
    local = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
  } catch (e) {
    console.warn("No se pudo obtener video+audio, se intenta solo video:", e.message);
    local = await navigator.mediaDevices.getUserMedia({ video: true });
  }

  local.getTracks().forEach(t => pc.addTrack(t, local));
  const video = document.getElementById("videoRemoto");
  video.srcObject = local;
  video.muted = true; // evita el eco de tu propia voz en tu vista previa

  const oferta = await pc.createOffer();
  await pc.setLocalDescription(oferta);
  streamWs.send(JSON.stringify({ tipo: "oferta", sdp: oferta }));
}

document.getElementById("btnTomarPalabra").addEventListener("click", async () => {
  const btn = document.getElementById("btnTomarPalabra");

  if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
    alert(
      "Este navegador no permite acceder a la cámara en esta dirección.\n\n" +
      "Los navegadores solo dan acceso a cámara/micrófono en HTTPS o en 'localhost'. " +
      `Estás entrando por '${location.hostname}', que no es seguro para esto — ` +
      "abrí la página como http://localhost:3000 en esta misma máquina, o serví el sitio con HTTPS."
    );
    return;
  }

  try {
    esEmisor = true;
    btn.disabled = true;
    btn.textContent = "TRANSMITIENDO…";
    await iniciarOferta();
  } catch (e) {
    esEmisor = false;
    btn.disabled = false;
    btn.textContent = "TOMAR LA PALABRA Y TRANSMITIR";
    alert("No se pudo acceder a la cámara/micrófono: " + e.message);
  }
});

/* =====================================================================
   Panel 3 — FINALIZADO: resultado
   ===================================================================== */
function mostrarPanelFinal() {
  marcarEstado("FINALIZADO");
  mostrarResultado(debateActual);
}

function mostrarResultado(d) {
  document.getElementById("panelFinal").style.display = "";
  const puntuacionA = d.puntuacion_a != null ? d.puntuacion_a : 0;
  const puntuacionB = d.puntuacion_b != null ? d.puntuacion_b : 0;
  const ganador = d.ganador_equipo;

  document.getElementById("finalA").textContent = puntuacionA.toFixed(1);
  document.getElementById("finalB").textContent = puntuacionB.toFixed(1);
  document.getElementById("ganadorTexto").innerHTML = ganador
    ? `Ganó el <span class="a">Equipo ${ganador.toUpperCase()}</span>`
    : "Resultado pendiente";
}

cargarDebate();
