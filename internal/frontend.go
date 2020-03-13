package internal

import (
	"html/template"
	"log"
	"net/http"

	"github.com/jcawley/kyma-app-connector/pkg/connector"
)

//IndexHandler -
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("IndexHandler")
	status := connector.GetConnectionStatus()

	type Status struct {
		ConnectionStatus string
	}
	pageData := Status{
		ConnectionStatus: status,
	}

	tmpl := template.Must(template.ParseFiles("../../assets/templates/index.html"))
	tmpl.Execute(w, pageData)

}
