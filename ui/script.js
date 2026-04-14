document.addEventListener("DOMContentLoaded", () => {
  fetch("/api/info")
    .then((res) => res.json())
    .then((data) => {
      const urlEl = document.getElementById("serverUrl");
      if (urlEl && data.url) {
        urlEl.textContent = data.url;
      }
    })
    .catch((err) => console.error("Error fetching info:", err));

  const displaySelector = document.getElementById("displaySelector");
  if (displaySelector) {
    fetch("/api/displays")
      .then((res) => res.json())
      .then((data) => {
        if (data && data.count > 1) {
          let html = "";
          for (let i = 0; i < data.count; i++) {
            const isActive = i === data.current ? "display-selector__btn--active" : "";
            html += `<button class="display-selector__btn ${isActive}" data-id="${i}">Экран ${i + 1}</button>`;
          }
          displaySelector.innerHTML = html;

          const buttons = displaySelector.querySelectorAll(".display-selector__btn");
          buttons.forEach((btn) => {
            btn.addEventListener("click", () => {
              const id = btn.getAttribute("data-id");
              fetch(`/api/set_display?id=${id}`).then(() => {
                buttons.forEach((b) => b.classList.remove("display-selector__btn--active"));
                btn.classList.add("display-selector__btn--active");
              });
            });
          });
        } else {
          displaySelector.style.display = "none";
        }
      })
      .catch((err) => console.error("Error fetching displays:", err));
  }
});
