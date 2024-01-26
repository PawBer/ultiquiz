package handlers

import (
	"net/http"

	"github.com/PawBer/ultiquiz/templates/pages"
	"github.com/julienschmidt/httprouter"
)

func GetIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	component := pages.Hello("World")
	component.Render(r.Context(), w)
}
