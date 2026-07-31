(function () {
  var storageKey = "theme";
  var root = document.documentElement;
  var toggle = document.getElementById("theme-toggle");

  function systemDark() {
    return window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches;
  }

  function applyTheme(theme) {
    if (theme === "dark" || theme === "light") {
      root.setAttribute("data-theme", theme);
      return;
    }
    root.removeAttribute("data-theme");
  }

  function currentTheme() {
    var stored = localStorage.getItem(storageKey);
    if (stored === "dark" || stored === "light") {
      return stored;
    }
    return systemDark() ? "dark" : "light";
  }

  applyTheme(localStorage.getItem(storageKey));

  if (toggle) {
    toggle.addEventListener("click", function () {
      var next = currentTheme() === "dark" ? "light" : "dark";
      localStorage.setItem(storageKey, next);
      applyTheme(next);
    });
  }

  if (window.matchMedia) {
    window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", function () {
      if (!localStorage.getItem(storageKey)) {
        applyTheme(null);
      }
    });
  }
})();
