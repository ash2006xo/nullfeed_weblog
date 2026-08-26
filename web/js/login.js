document.getElementById("login-form").addEventListener("submit", async (e) => {
  e.preventDefault();

  const username = document.getElementById("username").value.trim();
  const password = document.getElementById("password").value;
  const errorEl = document.getElementById("error-msg");
  errorEl.style.display = "none";

  try {
    const user = await api.login(username, password);
    setCurrentUsername(user.username);
    window.location.href = "/index.html";
  } catch (err) {
    errorEl.textContent = err.message;
    errorEl.style.display = "block";
  }
});