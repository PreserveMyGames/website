package web

import (
	"net/http"

	"github.com/PreserveMyGames/website/internal/constants"
)

func setHTMLCacheHeaders(w http.ResponseWriter) {
	w.Header().Set(constants.HeaderCacheControl, constants.PageCacheControl)
	w.Header().Set(constants.HeaderPragma, "no-cache")
	w.Header().Set(constants.HeaderExpires, "0")
}
