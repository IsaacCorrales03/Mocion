// Utilidades compartidas de interfaz: reloj de la barra de navegación,
// estado de sesión en el nav, y formato de fechas/horas.
function iniciarReloj(elId) {
  const el = document.getElementById(elId);
  if (!el) return;
  const tick = () => { el.textContent = new Date().toLocaleTimeString("es-CR", { hour12: false }); };
  tick(); setInterval(tick, 1000);
}

function pintarSesionEnNav() {
  const s = API.sesion();
  const cont = document.getElementById("navSesion");
  if (!cont) return;
  if (s) {
    cont.innerHTML = `<span>${s.nombre}</span><button id="btnSalir">SALIR</button>`;
    document.getElementById("btnSalir").addEventListener("click", () => {
      API.cerrarSesion(); location.href = "login.html";
    });
  } else {
    cont.innerHTML = `<a class="btn" href="login.html">INGRESAR</a>`;
  }
}

function requerirSesion() {
  const s = API.sesion();
  if (!s) { location.href = "login.html"; }
  return s;
}

function horaCorta(iso) {
  if (!iso) return "—";
  return new Date(iso).toLocaleTimeString("es-CR", { hour12: false, hour: "2-digit", minute: "2-digit" });
}

function fechaCorta(iso) {
  if (!iso) return "—";
  return new Date(iso).toLocaleDateString("es-CR", { day: "2-digit", month: "short" });
}
