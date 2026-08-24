const API_BASE = "/api";

async function apiFetch(path, options = {}) {
  const res = await fetch(API_BASE + path, {
    ...options,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    },
  });

  const isJson = res.headers.get("content-type")?.includes("application/json");
  const body = isJson ? await res.json() : null;

  if (!res.ok) {
    const message = body?.error || `Request failed (${res.status})`;
    throw new Error(message);
  }

  return body;
}

const api = {
  signup: (username, password) =>
    apiFetch("/signup", { method: "POST", body: JSON.stringify({ username, password }) }),

  login: (username, password) =>
    apiFetch("/login", { method: "POST", body: JSON.stringify({ username, password }) }),

  listBoards: () => apiFetch("/boards"),

  getBoard: (id) => apiFetch(`/boards/${id}`),

  createBoard: (data) =>
    apiFetch("/boards", { method: "POST", body: JSON.stringify(data) }),

  deleteBoard: (id) => apiFetch(`/boards/${id}`, { method: "DELETE" }),

  listComments: (boardId) => apiFetch(`/boards/${boardId}/comments`),

  createComment: (boardId, content) =>
    apiFetch(`/boards/${boardId}/comments`, {
      method: "POST",
      body: JSON.stringify({ content }),
    }),
};