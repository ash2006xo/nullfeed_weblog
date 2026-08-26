window.authReady.then((user) => {
  if (!user) window.location.href = "/login.html";
});

const createForm = document.getElementById("create-form");
const privateCheckbox = document.getElementById("is-private");
const sharingFields = document.getElementById("sharing-fields");
const imageInput = document.getElementById("image");
const imageStatus = document.getElementById("image-status");
const imagePreview = document.getElementById("image-preview");
let imageURL = null;

privateCheckbox.addEventListener("change", () => {
  sharingFields.classList.toggle("hidden", !privateCheckbox.checked);
});

imageInput.addEventListener("change", async () => {
  imageURL = null;
  imagePreview.innerHTML = "";
  const file = imageInput.files[0];
  if (!file) { imageStatus.textContent = ""; return; }
  if (file.size > 5 * 1024 * 1024) {
    imageInput.value = "";
    imageStatus.textContent = "Image must be 5 MB or smaller.";
    return;
  }

  const localURL = URL.createObjectURL(file);
  imagePreview.innerHTML = `<img src="${localURL}" alt="Preview">`;
  imageStatus.textContent = "Uploading image…";

  try {
    const result = await api.uploadImage(file);
    imageURL = result.url;
    imageStatus.textContent = "Image ready ✓";
  } catch (err) {
    imageInput.value = "";
    imagePreview.innerHTML = "";
    imageStatus.textContent = err.message;
  }
});

createForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  const errorEl = document.getElementById("error-msg");
  const submitButton = createForm.querySelector("button[type=submit]");
  errorEl.style.display = "none";

  const shareWith = document.getElementById("share-with").value.split(",").map((name) => name.trim()).filter(Boolean);
  submitButton.disabled = true;
  submitButton.innerHTML = "Publishing…";

  try {
    const board = await api.createBoard({
      title: document.getElementById("title").value.trim(),
      content: document.getElementById("content").value.trim(),
      image_url: imageURL,
      is_private: privateCheckbox.checked,
      share_with: privateCheckbox.checked ? shareWith : [],
    });
    window.location.href = `/weblog/${board.id}`;
  } catch (err) {
    errorEl.textContent = err.message;
    errorEl.style.display = "block";
  } finally {
    submitButton.disabled = false;
    submitButton.innerHTML = "Publish post <span>→</span>";
  }
});
