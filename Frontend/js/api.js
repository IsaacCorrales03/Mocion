// Cliente de la API de Moción. Ajusta BASE_URL al host donde corre el backend Go.
window.MOCION_API_URL = "https://mocion.onrender.com";
const API = (() => {
  const BASE_URL = window.MOCION_API_URL || `${location.protocol}//${location.hostname}:8080`;
  const WS_BASE = BASE_URL.replace(/^http/, "ws");

  function sesion() {
    try { return JSON.parse(localStorage.getItem("mocion_sesion")); }
    catch { return null; }
  }

  function guardarSesion(s) { localStorage.setItem("mocion_sesion", JSON.stringify(s)); }
  function cerrarSesion() { localStorage.removeItem("mocion_sesion"); }

  async function peticion(path, opciones = {}) {
    const res = await fetch(BASE_URL + path, {
      headers: { "Content-Type": "application/json" },
      ...opciones,
    });
    let cuerpo = null;
    try { cuerpo = await res.json(); } catch { /* sin cuerpo JSON */ }
    if (!res.ok) {
      const mensaje = (cuerpo && cuerpo.error) || `error ${res.status}`;
      throw new Error(mensaje);
    }
    return cuerpo;
  }

  return {
    BASE_URL, WS_BASE, sesion, guardarSesion, cerrarSesion,

    login: (email, password) => peticion("/login", { method: "POST", body: JSON.stringify({ email, password }) }),
    registrar: (nombre, email, password) => peticion("/register", { method: "POST", body: JSON.stringify({ nombre, email, password }) }),

    crearDebate: (titulo, descripcion, creador_id, rondas, participantes_por_equipo) =>
      peticion("/debates", { method: "POST", body: JSON.stringify({ titulo, descripcion, creador_id, rondas, participantes_por_equipo }) }),
    obtenerDebate: (id) => peticion(`/debates/${id}`),
    debatesEnVivo: () => peticion("/debates/en-vivo"),
    debatesProgramados: () => peticion("/debates/programados"),
    iniciarDebate: (id, forzar) => peticion(`/debates/${id}/iniciar`, { method: "POST", body: JSON.stringify({ forzar: !!forzar }) }),
    finalizarDebate: (id) => peticion(`/debates/${id}/finalizar`, { method: "POST" }),

    buscarUsuarios: (texto) => peticion(`/usuarios/buscar?q=${encodeURIComponent(texto)}`),
    asignarParticipante: (id, usuario_id, equipo) => peticion(`/debates/${id}/participantes`, { method: "POST", body: JSON.stringify({ usuario_id, equipo }) }),
    listarParticipantes: (id) => peticion(`/debates/${id}/participantes`),

    votar: (id, usuario_id, opcion) => peticion(`/debates/${id}/votar`, { method: "POST", body: JSON.stringify({ usuario_id, opcion }) }),
    resultados: (id, ronda) => peticion(`/debates/${id}/resultados?ronda=${ronda || 1}`),

    asignarJurado: (id, usuario_id) => peticion(`/debates/${id}/jurado`, { method: "POST", body: JSON.stringify({ usuario_id }) }),
    obtenerJurado: (id) => peticion(`/debates/${id}/jurado`),
    puntuar: (id, jurado_usuario_id, participante_usuario_id, puntuacion) =>
      peticion(`/debates/${id}/puntuacion`, { method: "POST", body: JSON.stringify({ jurado_usuario_id, participante_usuario_id, puntuacion }) }),
    obtenerPuntuaciones: (id) => peticion(`/debates/${id}/puntuaciones`),

    chatSocket: (id) => new WebSocket(`${WS_BASE}/debates/${id}/chat`),
    streamSocket: (id) => new WebSocket(`${WS_BASE}/debates/${id}/stream`),
    presenciaSocket: (id, usuarioId) => new WebSocket(`${WS_BASE}/debates/${id}/presencia?usuario_id=${usuarioId}`),
    conectados: (id) => peticion(`/debates/${id}/conectados`),
  };
})();
