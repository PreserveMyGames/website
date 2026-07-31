(function () {
  var form = document.querySelector("[data-blog-search-form]");
  var input = document.querySelector("[data-blog-search-input]");
  var list = document.getElementById("blog-post-list");
  if (!form || !input || !list) {
    return;
  }

  var status = document.getElementById("blog-search-status");
  var empty = document.getElementById("blog-search-empty");
  var items = Array.prototype.slice.call(list.querySelectorAll("li"));
  var matchOne = list.getAttribute("data-match-one") || "1 post matches";
  var matchMany = list.getAttribute("data-match-many") || "%d posts match";

  function normalize(value) {
    return value.toLowerCase().replace(/\s+/g, " ").trim();
  }

  function matchText(item, query) {
    var haystack = normalize(item.getAttribute("data-search") || item.textContent || "");
    var terms = query.split(/\s+/).filter(Boolean);
    for (var i = 0; i < terms.length; i++) {
      if (haystack.indexOf(terms[i]) === -1) {
        return false;
      }
    }
    return true;
  }

  function setStatus(message) {
    if (!status) {
      return;
    }
    if (!message) {
      status.hidden = true;
      status.textContent = "";
      return;
    }
    status.hidden = false;
    status.textContent = message;
  }

  function filterPosts() {
    var query = normalize(input.value);
    if (!query) {
      items.forEach(function (item) {
        item.classList.remove("is-hidden");
      });
      if (empty) {
        empty.hidden = true;
      }
      setStatus("");
      return;
    }

    var visible = 0;
    items.forEach(function (item) {
      var show = matchText(item, query);
      item.classList.toggle("is-hidden", !show);
      if (show) {
        visible += 1;
      }
    });

    if (empty) {
      empty.hidden = visible > 0;
    }

    if (visible === 1) {
      setStatus(matchOne);
      return;
    }
    setStatus(matchMany.replace("%d", String(visible)));
  }

  form.addEventListener("submit", function (event) {
    var query = normalize(input.value);
    if (!query) {
      return;
    }
    event.preventDefault();
    filterPosts();
  });

  input.addEventListener("input", filterPosts);
  input.addEventListener("search", function () {
    if (!input.value) {
      filterPosts();
    }
  });

  if (input.value.trim()) {
    filterPosts();
  }
})();
