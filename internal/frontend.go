package internal

import (
	"html/template"
	"log"
	"net/http"
)

//IndexHandler -
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("IndexHandler")
	tmpl := template.Must(template.ParseFiles("../../assets/templates/index.html"))
	tmpl.Execute(w, nil)
}
