package routes

import (
	"html/template"
	"net/http"
)

func (bs *BadgeServer) getTerms(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.ParseFS(badgeTemplate, "templates/terms.gohtml"))
	w.Header().Set("Content-Type", "text/html")
	err := tmpl.Execute(w, struct{}{})
	if err != nil {
		http.Error(w, "unable to template index", 500)
	}
}
