iniciarReloj("clock");
pintarSesionEnNav();

function tarjetaDebate(d, { enVivo }) {
  const div = document.createElement("a");
  div.className = "card";
  div.href = `debate.html#${d.id}`;
  div.innerHTML = `
    <div class="k">${enVivo ? '<span class="live-mark"><span class="live-dot"></span> EN DIRECTO</span>' : "PROGRAMADO · " + fechaCorta(d.creado_en)}</div>
    <h3>${d.titulo}</h3>
    <p>${d.descripcion || "Sin descripción adicional."}</p>
  `;
  return div;
}

async function cargar() {
  const listaVivo = document.getElementById("listaVivo");
  const listaProg = document.getElementById("listaProg");

  try {
    const vivos = await API.debatesEnVivo();
    document.getElementById("countVivo").textContent = `${vivos.length} ACTIVOS`;
    listaVivo.innerHTML = "";
    if (!vivos.length) {
      listaVivo.outerHTML = '<div class="empty" id="listaVivo">No hay debates en directo en este momento.</div>';
    } else {
      vivos.forEach(d => listaVivo.appendChild(tarjetaDebate(d, { enVivo: true })));
    }
  } catch (e) {
    listaVivo.outerHTML = `<div class="empty">No se pudo conectar con el servidor (${e.message}).</div>`;
  }

  try {
    const prog = await API.debatesProgramados();
    document.getElementById("countProg").textContent = `${prog.length} PROGRAMADOS`;
    listaProg.innerHTML = "";
    if (!prog.length) {
      listaProg.outerHTML = '<div class="empty" id="listaProg">No hay debates programados todavía.</div>';
    } else {
      prog.forEach(d => listaProg.appendChild(tarjetaDebate(d, { enVivo: false })));
    }
  } catch (e) {
    listaProg.outerHTML = `<div class="empty">No se pudo conectar con el servidor (${e.message}).</div>`;
  }
}

cargar();
