iniciarReloj("clock");

const tabLogin = document.getElementById("tabLogin");
const tabRegistro = document.getElementById("tabRegistro");
const formLogin = document.getElementById("formLogin");
const formRegistro = document.getElementById("formRegistro");
const titulo = document.getElementById("titulo");

tabLogin.addEventListener("click", () => {
  tabLogin.classList.add("on"); tabRegistro.classList.remove("on");
  formLogin.style.display = ""; formRegistro.style.display = "none";
  titulo.textContent = "Ingresar";
});
tabRegistro.addEventListener("click", () => {
  tabRegistro.classList.add("on"); tabLogin.classList.remove("on");
  formRegistro.style.display = ""; formLogin.style.display = "none";
  titulo.textContent = "Crear cuenta";
});

formLogin.addEventListener("submit", async (e) => {
  e.preventDefault();
  const err = document.getElementById("loginError");
  err.textContent = "";
  try {
    const r = await API.login(
      document.getElementById("loginEmail").value,
      document.getElementById("loginPassword").value
    );
    API.guardarSesion({ token: r.token, nombre: r.nombre, email: r.email, id: r.id });
    location.href = "index.html";
  } catch (e2) {
    err.textContent = e2.message;
  }
});

formRegistro.addEventListener("submit", async (e) => {
  e.preventDefault();
  const err = document.getElementById("regError");
  const ok = document.getElementById("regOk");
  err.textContent = ""; ok.textContent = "";
  try {
    const r = await API.registrar(
      document.getElementById("regNombre").value,
      document.getElementById("regEmail").value,
      document.getElementById("regPassword").value
    );
    ok.textContent = "Cuenta creada. Ahora podés iniciar sesión.";
    tabLogin.click();
    document.getElementById("loginEmail").value = r.email;
  } catch (e2) {
    err.textContent = e2.message;
  }
});
