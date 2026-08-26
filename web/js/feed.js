let allBoards = [];
let activeFilter = "all";

function renderBoardCard(board, index) {
  const privacyBadge = board.is_private
    ? `<span class="badge badge-private">Private</span>`
    : `<span class="badge badge-public">Public</span>`;
  const image = board.image_url
    ? `<div class="feed-card-image"><img src="${escapeHtml(board.image_url)}" alt=""></div>`
    : `<div class="feed-card-image feed-card-image-empty"><span>${escapeHtml(getInitials(board.title))}</span></div>`;
  const excerpt = escapeHtml(board.content.length > 150 ? `${board.content.slice(0, 150)}…` : board.content);

  return `
    <article class="feed-card" style="--delay:${Math.min(index * 45, 300)}ms">
      <a class="feed-card-link" href="/weblog/${board.id}">
        ${image}
        <div class="feed-card-body">
          <div class="feed-card-top">${privacyBadge}<span class="meta">${formatDate(board.created_at)}</span></div>
          <h2>${escapeHtml(board.title)}</h2>
          <p>${excerpt}</p>
          <div class="feed-card-footer">
            <span class="author-avatar">${escapeHtml(getInitials(board.author_name))}</span>
            <span>by ${escapeHtml(board.author_name)}</span>
            <span class="read-more">Read post →</span>
          </div>
        </div>
      </a>
    </article>
  `;
}

function filteredBoards() {
  if (activeFilter === "public") return allBoards.filter((board) => !board.is_private);
  if (activeFilter === "private") return allBoards.filter((board) => board.is_private);
  return allBoards;
}

function renderFeed() {
  const listEl = document.getElementById("feed-list");
  const query = document.getElementById("feed-search")?.value.trim().toLowerCase() || "";
  let boards = filteredBoards();

  if (query) {
    boards = boards.filter((board) =>
      `${board.title} ${board.content} ${board.author_name}`.toLowerCase().includes(query)
    );
  }

  if (!boards.length) {
    listEl.innerHTML = `
      <div class="empty-state card">
        <div class="empty-icon">⌁</div>
        <h2>${query ? "Nothing found" : "Your feed is quiet"}</h2>
        <p>${query ? "Try another search term." : "Be the first to leave something worth reading."}</p>
        ${getCurrentUsername() ? `<a class="btn" href="/create.html">Write a post <span>→</span></a>` : `<a class="btn" href="/signup.html">Join nullfeed <span>→</span></a>`}
      </div>
    `;
    return;
  }

  listEl.innerHTML = boards.map(renderBoardCard).join("");
}

function setupFilters() {
  document.querySelectorAll("[data-filter]").forEach((button) => {
    button.addEventListener("click", () => {
      activeFilter = button.dataset.filter;
      document.querySelectorAll("[data-filter]").forEach((item) => item.classList.remove("active"));
      button.classList.add("active");
      renderFeed();
    });
  });

  document.getElementById("feed-search")?.addEventListener("input", renderFeed);
}

async function loadFeed() {
  const listEl = document.getElementById("feed-list");
  try {
    await window.authReady;
    allBoards = await api.listBoards();
    const count = document.getElementById("post-count");
    if (count) count.textContent = `${allBoards.length} ${allBoards.length === 1 ? "post" : "posts"}`;
    setupFilters();
    renderFeed();
  } catch (err) {
    listEl.innerHTML = `<div class="empty-state card"><div class="empty-icon">!</div><h2>Couldn't load the feed</h2><p>${escapeHtml(err.message)}</p><button class="btn" onclick="loadFeed()">Try again <span>↻</span></button></div>`;
  }
}

loadFeed();
