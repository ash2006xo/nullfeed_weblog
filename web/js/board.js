function getBoardIdFromUrl() {
  const pathMatch = window.location.pathname.match(/^\/weblog\/(\d+)\/?$/);
  if (pathMatch) return pathMatch[1];
  return new URLSearchParams(window.location.search).get("id");
}

function renderComment(comment) {
  return `
    <div class="comment">
      <div class="comment-avatar">${escapeHtml(getInitials(comment.author_name))}</div>
      <div class="comment-body">
        <div class="meta"><strong>${escapeHtml(comment.author_name)}</strong> · ${formatDateTime(comment.created_at)}</div>
        <div>${escapeHtml(comment.content)}</div>
      </div>
    </div>
  `;
}

async function loadBoard() {
  await window.authReady;
  const id = getBoardIdFromUrl();
  const detailEl = document.getElementById("board-detail");

  if (!id) {
    detailEl.innerHTML = `<div class="empty-state card"><div class="empty-icon">?</div><h2>Post not found</h2><p>No post was specified.</p><a class="btn" href="/">Back to feed</a></div>`;
    return;
  }

  try {
    const board = await api.getBoard(id);
    const currentUsername = getCurrentUsername();
    const isOwner = currentUsername && currentUsername === board.author_name;
    const imageHtml = board.image_url
      ? `<div class="board-cover"><img class="board-image" src="${escapeHtml(board.image_url)}" alt=""></div>`
      : "";
    const privacyBadge = board.is_private ? `<span class="badge badge-private">Private</span>` : `<span class="badge badge-public">Public</span>`;
    const deleteButtonHtml = isOwner ? `<button class="danger" id="delete-btn">Delete post</button>` : "";

    detailEl.innerHTML = `
      <a class="back-link" href="/">← Back to feed</a>
      <article class="board-article">
        <div class="board-header">
          <div class="board-eyebrow">${privacyBadge}<span>${formatDate(board.created_at)}</span></div>
          <h1>${escapeHtml(board.title)}</h1>
          <div class="board-author"><span class="author-avatar large">${escapeHtml(getInitials(board.author_name))}</span><span>Written by <strong>${escapeHtml(board.author_name)}</strong></span></div>
        </div>
        ${imageHtml}
        <div class="board-content">${escapeHtml(board.content)}</div>
        ${isOwner ? `<div class="board-actions">${deleteButtonHtml}</div>` : ""}
      </article>

      ${isOwner && board.is_private ? `
        <section id="share-management" class="card share-card">
          <div class="section-heading"><div><span class="section-kicker">Private access</span><h2>Who can read this?</h2></div><span class="share-lock">⌘</span></div>
          <p class="meta">Only you and the people listed here can open this post.</p>
          <div class="share-input-row"><input id="share-usernames" placeholder="alice, bob" autocomplete="off"><button id="save-shares" type="button">Update access</button></div>
          <p class="error" id="share-error" style="display:none;"></p>
        </section>` : ""}

      <section class="comments-section">
        <div class="section-heading"><div><span class="section-kicker">Discussion</span><h2>Comments</h2></div><span id="comment-count" class="comment-count"></span></div>
        <div id="comments-list" class="comments-list"><div class="skeleton skeleton-comment"></div><div class="skeleton skeleton-comment"></div></div>
        <div id="comment-form-container"></div>
      </section>
    `;

    if (isOwner) {
      document.getElementById("delete-btn")?.addEventListener("click", async () => {
        if (!confirm("Delete this post permanently?")) return;
        try {
          await api.deleteBoard(id);
          window.location.href = "/";
        } catch (err) {
          alert("Failed to delete: " + err.message);
        }
      });
    }

    if (isOwner && board.is_private) await loadShares(id);
    await loadComments(id);
    renderCommentForm(id);
  } catch (err) {
    detailEl.innerHTML = `<div class="empty-state card"><div class="empty-icon">!</div><h2>We couldn't open this post</h2><p>${escapeHtml(err.message)}</p><a class="btn" href="/">Back to feed</a></div>`;
  }
}

async function loadShares(boardId) {
  try {
    const result = await api.listShares(boardId);
    const input = document.getElementById("share-usernames");
    if (input) input.value = result.usernames.join(", ");
    document.getElementById("save-shares")?.addEventListener("click", async () => {
      const errorEl = document.getElementById("share-error");
      errorEl.style.display = "none";
      const usernames = input.value.split(",").map((name) => name.trim()).filter(Boolean);
      try {
        await api.replaceShares(boardId, usernames);
        const button = document.getElementById("save-shares");
        button.textContent = "Saved ✓";
        setTimeout(() => { button.textContent = "Update access"; }, 1500);
      } catch (err) {
        errorEl.textContent = err.message;
        errorEl.style.display = "block";
      }
    });
  } catch (err) {
    const errorEl = document.getElementById("share-error");
    if (errorEl) { errorEl.textContent = err.message; errorEl.style.display = "block"; }
  }
}

async function loadComments(boardId) {
  const listEl = document.getElementById("comments-list");
  try {
    const comments = await api.listComments(boardId);
    const count = document.getElementById("comment-count");
    if (count) count.textContent = `${comments.length} ${comments.length === 1 ? "comment" : "comments"}`;
    listEl.innerHTML = comments.length ? comments.map(renderComment).join("") : `<div class="comments-empty">No comments yet. Start the conversation.</div>`;
  } catch (err) {
    listEl.innerHTML = `<p class="error">Failed to load comments: ${escapeHtml(err.message)}</p>`;
  }
}

function renderCommentForm(boardId) {
  const container = document.getElementById("comment-form-container");
  const username = getCurrentUsername();
  if (!username) {
    container.innerHTML = `<div class="login-prompt"><span>Want to join the conversation?</span><a href="/login.html">Log in to comment →</a></div>`;
    return;
  }

  container.innerHTML = `
    <form id="comment-form" class="comment-form">
      <div class="comment-composer-avatar">${escapeHtml(getInitials(username))}</div>
      <div class="comment-composer-main">
        <textarea id="comment-content" placeholder="Leave a thoughtful comment..." maxlength="5000" required></textarea>
        <div class="comment-form-footer"><span class="meta">Posting as ${escapeHtml(username)}</span><button type="submit">Comment <span>→</span></button></div>
        <p class="error" id="comment-error" style="display:none;"></p>
      </div>
    </form>
  `;

  document.getElementById("comment-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const contentEl = document.getElementById("comment-content");
    const errorEl = document.getElementById("comment-error");
    const content = contentEl.value.trim();
    errorEl.style.display = "none";
    if (!content) return;
    try {
      await api.createComment(boardId, content);
      contentEl.value = "";
      await loadComments(boardId);
    } catch (err) {
      errorEl.textContent = err.message;
      errorEl.style.display = "block";
    }
  });
}

loadBoard();
