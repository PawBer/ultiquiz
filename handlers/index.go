package handlers

import (
	"net/http"

	"github.com/PawBer/ultiquiz/templates/pages"
)

func GetIndex(w http.ResponseWriter, r *http.Request) {
	component := pages.Hello("World")
	component.Render(r.Context(), w)
}
