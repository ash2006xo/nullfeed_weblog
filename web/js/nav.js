function getCurrentUsername() {
  return sessionStorage.getItem("username");
}

function setCurrentUsername(username) {
  sessionStorage.setItem("username", username);
}

function clearCurrentUsername() {
  sessionStorage.removeItem("username");
}

function renderNav() {
  const username = getCurrentUsername();
  const linksEl = document.getElementById("nav-links");
  if (!linksEl) return;

  if (username) {
    linksEl.innerHTML = `
      <a href="/create.html">New post</a>
      <span>${username}</span>
      <button id="logout-btn">Log out</button>
    `;
    document.getElementById("logout-btn").addEventListener("click", () => {
      clearCurrentUsername();
      window.location.href = "/index.html";
    });
  } else {
    linksEl.innerHTML = `
      <a href="/login.html">Log in</a>
      <a href="/signup.html">Sign up</a>
    `;
  }
}

document.addEventListener("DOMContentLoaded", renderNav);