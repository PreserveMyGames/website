(function () {
  var wrap = document.querySelector("[data-locale-menu]");
  if (!wrap) {
    return;
  }

  var trigger = wrap.querySelector("[data-locale-trigger]");
  var panel = wrap.querySelector("[data-locale-panel]");
  if (!trigger || !panel) {
    return;
  }

  function open() {
    trigger.setAttribute("aria-expanded", "true");
    panel.hidden = false;
  }

  function close() {
    trigger.setAttribute("aria-expanded", "false");
    panel.hidden = true;
  }

  function isOpen() {
    return trigger.getAttribute("aria-expanded") === "true";
  }

  trigger.addEventListener("click", function (event) {
    event.stopPropagation();
    if (isOpen()) {
      close();
      return;
    }
    open();
  });

  document.addEventListener("click", function (event) {
    if (!isOpen()) {
      return;
    }
    if (!wrap.contains(event.target)) {
      close();
    }
  });

  document.addEventListener("keydown", function (event) {
    if (event.key === "Escape" && isOpen()) {
      close();
      trigger.focus();
    }
  });

  panel.addEventListener("click", function () {
    close();
  });
})();
