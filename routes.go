package main

import (
	"fmt"
	"html/template"
	"net/http"
	_ "net/url"
	"os"
	_ "path"
	"path/filepath"
	"github.com/gorilla/sessions"

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

func listHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch new store.

	// Get or create sessions (both short and long lived)
	session1m, err := store.Get(r, "session1m")
	if_err_panic("23gh34: error creating session with 1m expiry", err)
	session1m.Options.MaxAge = 60 //seconds

	session5m, err := store.Get(r, "session5m")
	if_err_panic("12gh34: error creating session with 5m expiry", err)
	session5m.Options.MaxAge = 5*60 //seconds

	session10m, err := store.Get(r, "session10m")
	if_err_panic("df34sv: error creating session with 10m expiry", err)
	session10m.Options.MaxAge = 10*60 //seconds

	// Add a value
	session1m.Values["1m"] = "means 1 min"
	//	session5m.Values["5m"] = "means 5 min"
	//	session10m.Values["10m"] = "means 10 min"

	if str, ok := session1m.Values["1m"].(string); ok {
		fmt.Printf("Returned:" + str + "\n")
	} else {
		fmt.Printf("Type assert failed(1m)\n")
	}

	if str2, ok := session5m.Values["5m"].(string); ok {
		fmt.Printf("Returned:" + str2 + "\n")
	} else {
		fmt.Printf("Type assert failed(5m)\n")
	}

	if str3, ok := session10m.Values["10m"].(string); ok {
		fmt.Printf("Returned:" + str3 + "\n")
	} else {
		fmt.Printf("Type assert failed(10m)\n")
	}

	// Delete session.
	//	session.Options.MaxAge = -1

	// Save.
	err = sessions.Save(r, w);
	if_err_panic("12ds34: error saving session", err)

	// prepare a statement
	stmt, err := db.Prepare("SELECT * FROM items")
	if_err_panic("57HNB: malformed statement", err)
	defer stmt.Close()

	rows, err := stmt.Query()
	if_err_panic("12SDC: failed statement", err)

	var Items []Item
	var scanItem Item

	for rows.Next() {
		err := rows.Scan(&scanItem.Id, &scanItem.Ke_user_id, &scanItem.MinecraftId, &scanItem.Category, &scanItem.Name, &scanItem.Found, &scanItem.I1, &scanItem.I2, &scanItem.I3, &scanItem.I4, &scanItem.I5, &scanItem.I6, &scanItem.I7, &scanItem.I8, &scanItem.I9, &scanItem.notified_obsolete, &scanItem.notified_wrong, &scanItem.has_image, &scanItem.current, &scanItem.date_created)
		if_err_panic("23DFV: error during scan", err)
		//fmt.Println(id, name)
		Items = append(Items, scanItem)
	}

	rows.Close()

	fmt.Println("%#+v", Items)

	listContext := ListContext{"", false, User{"1", "me@you.com"}, Items}

	if r.URL.Query().Get("debug") == "true" {
		fmt.Fprintf(w, "--- CONTEXT---\n")
		fmt.Fprintf(w, "%#+v\n", listContext)
		fmt.Fprintf(w, "--- GENERATED PAGE ---\n")
	}

	err = allTemplates.ExecuteTemplate(w, "home", listContext)
	if_err_panic("error executing template: ", err)
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
