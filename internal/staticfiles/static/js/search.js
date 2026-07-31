(function () {
  var container = document.getElementById("search-results");
  if (!container || typeof Fuse === "undefined") {
    return;
  }

  var input = document.getElementById("q");
  var indexURL = container.getAttribute("data-index-url");
  var emptyMsg = container.getAttribute("data-empty") || "No results found.";
  var fuse = null;

  function render(results) {
    if (!results.length) {
      container.innerHTML = "<p>" + emptyMsg + "</p>";
      return;
    }

    var html = "<ul class=\"result-list\">";
    results.forEach(function (item) {
      var entry = item.item;
      html += "<li><a href=\"" + entry.url + "\">" + entry.title + "</a>";
      if (entry.description) {
        html += "<p>" + entry.description + "</p>";
      }
      html += "</li>";
    });
    html += "</ul>";
    container.innerHTML = html;
  }

  function search(query) {
    if (!fuse || !query) {
      return;
    }
    render(fuse.search(query));
  }

  fetch(indexURL)
    .then(function (response) {
      return response.json();
    })
    .then(function (data) {
      fuse = new Fuse(data, {
        keys: ["title", "description", "body", "tags"],
        threshold: 0.35
      });
      if (input && input.value) {
        search(input.value.trim());
      }
    })
    .catch(function () {
      return;
    });

  if (input) {
    input.addEventListener("input", function () {
      search(input.value.trim());
    });
  }
})();
