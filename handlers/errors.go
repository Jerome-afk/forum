package handlers

import (
	"fmt"
	"net/http"
	// "strings"
)

// StaticMiddleware handles access to static files
func StaticMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := "/" + r.URL.Path

		// Block direct access to /static/, /static/css/, /static/js/
		if path == "/" || path == "/css/" || path == "/js/" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Println("Error")
			renderTemplate(w, "404.html", nil)
			return
		}

		// Otherwise, serve the file
		next.ServeHTTP(w, r)
	})
}

func NoPageHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	renderTemplate(w, "404.html", nil)
}

func ServerProblemHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	renderTemplate(w, "500.html", nil)
}
