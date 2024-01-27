package handlers

import "net/http"

func (app *Application) LogRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.InfoLog.Printf("%s %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.RequestURI)

		h.ServeHTTP(w, r)
	})
}
