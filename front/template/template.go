package template

import (
	"html/template"
	"net/http"
)

func Render(w http.ResponseWriter, name string, content interface{}) {
	t := template.Must(template.ParseFiles(
		"template/layout.html", "template/header.html", "template/"+name))
	if err := t.ExecuteTemplate(w, "layout", content); err != nil {
		panic(err)
	}
}
