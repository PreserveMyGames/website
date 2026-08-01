(function () {
  var toggle = document.getElementById("nav-drawer-toggle");
  if (!toggle) {
    return;
  }

  var drawerLinks = document.querySelectorAll(".site-nav--drawer a");
  drawerLinks.forEach(function (link) {
    link.addEventListener("click", function () {
      toggle.checked = false;
    });
  });

  document.addEventListener("keydown", function (event) {
    if (event.key === "Escape" && toggle.checked) {
      toggle.checked = false;
    }
  });
})();
