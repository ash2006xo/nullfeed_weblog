let currentUser = null;

function getCurrentUsername() { return currentUser?.username || null; }
function setCurrentUsername(username) { currentUser = username ? { username } : null; }
function clearCurrentUsername() { currentUser = null; }

async function loadCurrentUser() {
  try { currentUser = await api.me(); } catch (_) { currentUser = null; }
  renderNav();
  return currentUser;
}

function renderNav() {
  const linksEl = document.getElementById("nav-links");
  if (!linksEl) return;
  if (currentUser) {
    linksEl.innerHTML = `
      <a class="nav-create" href="/create.html"><span>＋</span> New post</a>
      <span class="nav-user"><span class="author-avatar">${escapeHtml(getInitials(currentUser.username))}</span>${escapeHtml(currentUser.username)}</span>
      <button class="nav-logout" id="logout-btn" type="button">Log out</button>
    `;
    document.getElementById("logout-btn").addEventListener("click", async () => {
      const button = document.getElementById("logout-btn");
      button.disabled = true;
      try { await api.logout(); } catch (_) {}
      clearCurrentUsername();
      window.location.href = "/";
    });
  } else {
    linksEl.innerHTML = `<a href="/login.html">Log in</a><a class="nav-signup" href="/signup.html">Sign up</a>`;
  }
}

window.authReady = loadCurrentUser();
