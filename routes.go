package main

import (
	"fmt"
	"net/http"
	"os"
	"github.com/gorilla/sessions"

	"github.com/gorilla/schema"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/sendgrid/sendgrid-go"

)


func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch new store.
	log.Notice("/home")
	// sessions & user management ///////////////////////////
	session30m, session1y := getSessions(r)
	userState, extendedState := assertUser(w, r, session30m, session1y)

	log.Debug("userState is: %#+v", userState)
	log.Debug("extendedState is: %#+v", extendedState)

	// context & templates ///////////////////////////////////
	var contextFlashes []string

	for _, flash := range session30m.Flashes() {
		if str, ok := flash.(string); ok {
			contextFlashes = append(contextFlashes, str)
		}
	}

	// call this BEFORE writing to the http.ResponseWriter
	setSessions(w, r)

	// Delete session.
	//	session.Options.MaxAge = -1

	// Save.
	err := sessions.Save(r, w);
	if_err_panic("12ds34: error saving session", err)

	// prepare a statement
	stmt, err := db.Prepare("SELECT * FROM items")
	if_err_panic("57HNB: malformed statement", err)
	defer stmt.Close()

	rows, err := stmt.Query()
	if_err_panic("12SDC: failed statement", err)

	var items []Item
	var scanItem Item

	for rows.Next() {
		err := rows.Scan(&scanItem.Id, &scanItem.Ke_user_id, &scanItem.MinecraftId, &scanItem.Category, &scanItem.Name, &scanItem.Found, &scanItem.I1, &scanItem.I2, &scanItem.I3, &scanItem.I4, &scanItem.I5, &scanItem.I6, &scanItem.I7, &scanItem.I8, &scanItem.I9, &scanItem.notified_obsolete, &scanItem.notified_wrong, &scanItem.has_image, &scanItem.current, &scanItem.date_created)
		if_err_panic("23DFV: error during scan", err)
		//fmt.Println(id, name)
		items = append(items, scanItem)
	}

	rows.Close()

	fmt.Println("%#+v", items)

	listContext := ListContext{"", false, User{"1", "me@you.com"}, items, userState}

	if r.URL.Query().Get("debug") == "true" {
		fmt.Fprintf(w, "--- CONTEXT---\n")
		fmt.Fprintf(w, "%#+v\n", listContext)
		fmt.Fprintf(w, "--- GENERATED PAGE ---\n")
	}

	err = allTemplates.ExecuteTemplate(w, "list", listContext)
	if_err_panic("error executing template: ", err)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	log.Notice("/login")

	// sessions & user management ///////////////////////////
	session30m, session1y := getSessions(r)
	userState, extendedState := assertUser(w, r, session30m, session1y)

	// context & templates ///////////////////////////////////
	var contextFlashes []string

	for _, flash := range session30m.Flashes() {
		if str, ok := flash.(string); ok {
			contextFlashes = append(contextFlashes, str)
		}
	}




	// we create a translator using either userState (if logged)
	// or extendedState


	// call this BEFORE writing to the http.ResponseWriter
	setSessions(w, r)

	context := Context{r, userState, extendedState, contextFlashes}
	//	fmt.Printf("%#+v", context)
	//	goon.Dump(context)

	if r.URL.Query().Get("debug") == "true" {
		fmt.Fprintf(w, "--- CONTEXT---\n")
		fmt.Fprintf(w, "%#+v\n", context)
		fmt.Fprintf(w, "--- GENERATED PAGE ---\n")
	}

	// mutex rlock LanguageAtlas before passing to the template that will read it


	err := allTemplates.ExecuteTemplate(w, "login", context)
	if_err_panic("error executing template: ", err)
}

func authInternalLoginHandler(w http.ResponseWriter, r *http.Request) {
	// this is an internal handler that does not generate a page
	// it performs action altering some store values, then redirects

	log.Notice("/auth/internal/login")
	// sessions & user management ///////////////////////////
	session30m, session1y := getSessions(r)

	var userState UserState
	var extendedState ExtendedState
	var loginCredentials LoginCredentials

	// type assert userState and extendedState
	// or keep the ones you got
	var ok bool
	if userState, ok = session30m.Values["userState"].(UserState); ok {
		// RTA ok
	} else {
		// RTA failed, default value is ok
	}

	if extendedState, ok = session1y.Values["extendedState"].(ExtendedState); ok {
		// RTA ok
	} else {
		// RTA failed, default value is ok
	}

	// parse the forms
	err := r.ParseForm()
	if_err_panic("34ef56: Parse form failed", err)

	// create the gorilla/scheme decoder
	decoder := schema.NewDecoder()
	decoder.ZeroEmpty(true)

	// get email and password from form
	err = decoder.Decode(&loginCredentials, r.PostForm)
	if_err_panic("45ej59: Deconding failed", err)

	// get user record with that email of fail
	// scan a User record to check if it exists and if it's active
	// prepare the statement
	stmt, err := db.Prepare("SELECT user_id, nick, email, password, lang, lat, lng, rng, image, status, message, ftsync FROM users WHERE enabled = true AND activation_code = 0 AND email = ? LIMIT 1")
	if_err_panic("35H67: malformed statement", err)
	defer stmt.Close()

	// perform the query
	rows, err := stmt.Query(loginCredentials.Email)
	if_err_panic("17KBB: failed statement", err)

	// create a temp UserRow
	var scanUser UserRow

	for rows.Next() {
		// scan record inside scanUser's fields
		err := rows.Scan(&scanUser.UserID, &scanUser.Nick, &scanUser.Email, &scanUser.Password, &scanUser.Lang)
		scanUser.Enabled = true
		if_err_panic("67DFV: error during scan", err)
	}
	defer rows.Close()

	// get the encrypted password and check if match
	err = bcrypt.CompareHashAndPassword([]byte(scanUser.Password), []byte(loginCredentials.Password))
	if err == nil {
		// nil error means password match
		// if everything is fine populate the userState
		userState.Logged = true
		userState.UserID = scanUser.UserID
		userState.Nick = scanUser.Nick
		userState.Email = scanUser.Email
		userState.Lang = scanUser.Lang

		userState.RememberMe = loginCredentials.RememberMe
		session30m.Values["userState"] = userState
		log.Debug("userState = %#+v", userState)
		setSessions(w, r)
		http.Redirect(w, r, "/", 302)
	} else {
		// match FAIL
		// we don't alter the session userstate
		extendedState.LastLogin = loginCredentials.Email
		session30m.AddFlash("Error: wrong credentials, no such user or user not activated")
		session1y.Values["extendedState"] = extendedState
		setSessions(w, r)
		http.Redirect(w, r, "/login", 302)
	}
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
