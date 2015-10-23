package main

import (
	"fmt"
	"html/template"
	"net/http"
	_ "net/url"
	"os"
	_ "path"
	"path/filepath"
)

// PACKAGE GLOBALS /////////////////////////////////////////////////////////////////////////////////////////////////////

var allTemplates = template.New("home")
var templates []string

// INIT ////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	var err error
	basepath := "/var/go/craftbase/src/craftbase/templates/"

	filepath.Walk(basepath, addTemplate)

	//fp := path.Join(basepath, "*.tmpl")

	allTemplates, err = template.ParseFiles(templates...)
	checkErr("error parsing template: ", err)

	log.Info("routes.go: init completed\n")
}

// FUNCTIONS ////////////////////////////////////////////////////////////////////////////////////////////////////////

func homeHandler(w http.ResponseWriter, r *http.Request) {
	context := Context{"Potente Jedi", "Yoda"}

	if r.URL.Query().Get("debug") == "true" {
		fmt.Fprintf(w, "--- CONTEXT---\n")
		fmt.Fprintf(w, "%+v\n", context)
		fmt.Fprintf(w, "--- GENERATED PAGE ---\n")
	}

	/*fmt.Printf("%v\n", r.URL.String())
	fmt.Printf("%+v\n", params)
	fmt.Printf("%+v\n", context)*/

	err := allTemplates.ExecuteTemplate(w, "home", context)
	checkErr("error executing template: ", err)
}

func articlesHandler(w http.ResponseWriter, r *http.Request) {
	context := Context{"Potente Jedi", "Yoda"}
	err := allTemplates.ExecuteTemplate(w, "home", context)
	checkErr("error executing template: ", err)
}

func productsHandler(w http.ResponseWriter, r *http.Request) {
	context := Context{"Potente Jedi", "Yoda"}
	err := allTemplates.ExecuteTemplate(w, "products", context)
	checkErr("error executing template: ", err)
}

func checkErr(message string, err error) {
	if err != nil {
		log.Error(message, err)
	}
}

func addTemplate(path string, fi os.FileInfo, err error) error {
	if !fi.IsDir() {
		log.Info(path)
		templates = append(templates, path)
	}
	return err
}
