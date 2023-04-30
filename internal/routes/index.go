package routes

import (
	"html/template"
	"net/http"
)

func (bs *BadgeServer) getIndex(w http.ResponseWriter, _ *http.Request) {
	tmpl := template.Must(template.ParseFS(badgeTemplate, "templates/index.gohtml", "templates/head.gohtml", "templates/footer.gohtml"))
	w.Header().Set("Content-Type", "text/html")
	err := tmpl.Execute(w, struct {
		Title string
	}{"Create your last played badge"})
	if err != nil {
		http.Error(w, "unable to template index", 500)
	}
}
