(() => {
  "use strict";
  const themeButton = document.querySelector(".theme-button");
  const applyTheme = (theme) => {
    document.documentElement.dataset.theme = theme;
    themeButton.setAttribute("aria-label", "Switch to " + (theme === "dark" ? "light" : "dark") + " theme");
  };
  let storedTheme;
  try { storedTheme = localStorage.getItem("chessongo-theme"); } catch {}
  applyTheme(storedTheme === "light" ? "light" : "dark");
  themeButton.hidden = false;
  themeButton.addEventListener("click", () => {
    const theme = document.documentElement.dataset.theme === "dark" ? "light" : "dark";
    applyTheme(theme);
    try { localStorage.setItem("chessongo-theme", theme); } catch {}
  });
  document.querySelectorAll("[data-copy]").forEach(button => {
    button.hidden = false;
    button.addEventListener("click", async () => {
      const code = button.closest(".codeblock").querySelector("code").textContent;
      try {
        await navigator.clipboard.writeText(code);
        button.textContent = "Copied";
        document.getElementById("copy-status").textContent = "Code copied to clipboard.";
        setTimeout(() => { button.textContent = "Copy"; }, 1800);
      } catch {
        document.getElementById("copy-status").textContent = "Copy unavailable. Select the code and copy it manually.";
      }
    });
  });
  const data = document.getElementById("benchmark-data");
  const select = document.getElementById("workload");
  if (data && select) {
    const rows = JSON.parse(data.textContent);
    select.closest(".chart-control").hidden = false;
    select.addEventListener("change", () => {
      const row = rows.find(item => item.id === select.value);
      const max = Math.max(row.ours.ns, row.other.ns);
      document.getElementById("bar-ours").style.width = (row.ours.ns / max * 100) + "%";
      document.getElementById("bar-other").style.width = (row.other.ns / max * 100) + "%";
      document.getElementById("time-ours").textContent = row.ours.formatted;
      document.getElementById("time-other").textContent = row.other.formatted;
      document.getElementById("chart-result").textContent = row.result;
      document.getElementById("chart-note").textContent = row.note;
      document.getElementById("chart-title").textContent = row.title;
    });
  }
  const searchData = document.getElementById("search-data");
  const search = document.getElementById("docs-search");
  if (search && searchData) {
    const pages = JSON.parse(searchData.textContent);
    const results = document.getElementById("search-results");
    search.closest(".search-box").hidden = false;
    search.addEventListener("input", () => {
      const words = search.value.trim().toLowerCase().split(/\s+/).filter(Boolean);
      results.replaceChildren();
      if (!words.length) return;
      const matches = pages.filter(p => words.every(word => (p.title + " " + p.text).toLowerCase().includes(word))).slice(0, 7);
      if (!matches.length) {
        const item = document.createElement("li");
        item.textContent = "No matching pages.";
        results.append(item);
      }
      for (const match of matches) {
        const item = document.createElement("li");
        const a = document.createElement("a");
        a.href = match.href;
        a.textContent = match.title;
        item.append(a);
        results.append(item);
      }
    });
  }
})();
