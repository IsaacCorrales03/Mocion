iniciarReloj("clock");
pintarSesionEnNav();
const sesion = requerirSesion();

document.getElementById("formCrear").addEventListener("submit", async (e) => {
  e.preventDefault();
  const err = document.getElementById("crearError");
  const ok = document.getElementById("crearOk");
  err.textContent = ""; ok.textContent = "";
  try {
    const d = await API.crearDebate(
      document.getElementById("titulo").value,
      document.getElementById("descripcion").value,
      sesion.id,
      parseInt(document.getElementById("rondas").value, 10),
      parseInt(document.getElementById("participantesXEquipo").value, 10)
    );
    ok.textContent = "Debate creado. Ahora armá los equipos y el jurado…";
    setTimeout(() => location.href = `debate.html#${d.id}`, 700);
  } catch (e2) {
    err.textContent = e2.message;
  }
});
