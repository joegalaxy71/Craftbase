package main

import (
	"fmt"
	"net/http"
	"os"
	"math/rand"
	"strconv"
	"bytes"
	"time"
	"net/mail"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/schema"
	"github.com/sendgrid/sendgrid-go"
	"golang.org/x/crypto/bcrypt"
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
	err := sessions.Save(r, w)
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
func signUpHandler(w http.ResponseWriter, r *http.Request) {
	log.Notice("/signup")
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

	err := allTemplates.ExecuteTemplate(w, "signup", context)
	if_err_panic("error executing template: ", err)
}
func authInternalSignUpHandler(w http.ResponseWriter, r *http.Request) {
	// this is an internal handler that does not generate a page
	// it performs action altering some store values, then redirects

	log.Notice("/auth/internal/signup")
	// sessions & user management ///////////////////////////
	session30m, _ := getSessions(r)


	var signUpCredentials SignUpCredentials

	// parse the forms
	err := r.ParseForm()
	if_err_panic("33ef69: Parse form failed", err)

	// create the gorilla/scheme decoder
	decoder := schema.NewDecoder()
	decoder.ZeroEmpty(true)

	// decode signup struct from form
	err = decoder.Decode(&signUpCredentials, r.PostForm)
	if_err_panic("47ej60: Decoding failed", err)

	// check if we have a proper email address and password
	result, err := mail.ParseAddress(signUpCredentials.Email)
	signUpCredentials.Email = result.Address
	log.Debug("signUpCredentials=%+#v", signUpCredentials)

	// if no error we suppose to have a correct email address
	// parseAddress error is nil on sucess
	if err == nil {

		stmt, err := db.Prepare("SELECT user_id FROM users WHERE email = ? OR nick = ?;")
		if_err_panic("35H66: malformed statement", err)
		defer stmt.Close()

		// create a receiving structure for scan
		//var scanUser UserRow

		// execute the query and hope to get no result
		rows, err := stmt.Query(signUpCredentials.Email, signUpCredentials.Nick)
		defer rows.Close()

		// no result is good, it means the email and nick will be unique
		if !rows.Next() {

			// we create a new user record to be activated
			// prepare the statement

			stmt, err = db.Prepare("INSERT INTO users SET user_id = ?, kind = 'IN', role = 'U', nick = ?, email = ?, password = ?, activation_code = ?, recovery_code = 0;")
			if_err_panic("35H67: malformed statement", err)
			defer stmt.Close()

			// generate a new userID
			rand.Seed(time.Now().Unix())
			userID := minimapID()
			activationCode := minimapID()
			// encrypt the password
			encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(signUpCredentials.Password), 8)
			if_err_panic("23hy67: error encrypting password", err)
			// perform the query
			_, err = stmt.Exec(userID, signUpCredentials.Nick, signUpCredentials.Email, encryptedPassword, activationCode)
			if_err_panic("17KBB: failed exec", err)



			// prepare an EmailContext
			emailContext := EmailContext{r, signUpCredentials.Nick, signUpCredentials.Email, strconv.FormatInt(activationCode, 10)}

			// execute email template for HTML email part
			var msgHtml bytes.Buffer
			err = allTemplates.ExecuteTemplate(&msgHtml, "signupEmailHtml", emailContext)
			if_err_panic("MAIL01tmplHTML: error executing template: ", err)

			// execute email template for HTML email part
			var msgText bytes.Buffer
			err = allTemplates.ExecuteTemplate(&msgText, "signupEmailText", emailContext)
			if_err_panic("MAIL01tmplTEXT: error executing template: ", err)

			// send email
			sg := sendgrid.NewSendGridClient("ilnerdchuck", "Numero98")
			message := sendgrid.NewMail()
			message.AddTo(signUpCredentials.Email)
			message.AddToName(signUpCredentials.Nick)
			message.SetSubject("craftbase account activation")
			message.SetHTML(msgHtml.String())
			message.SetText(msgText.String())
			message.SetFrom("register@craftbase.com")
			if r := sg.Send(message); r == nil {
				log.Debug("Email sent!")
			} else {
				fmt.Println(r)
			}

			s := msgHtml.String()
			log.Debug("MailBody: %s", s)

			// user record succesfully created
			session30m.AddFlash("Success: user correctly created. Proceed to activation via the link we sent you trough email.")
			setSessions(w, r)
			http.Redirect(w, r, "/", 302)
		} else {
			// we have already a record with that email address, so we flash a message and redirect back to the signup
			session30m.AddFlash("Error: email of nickname are taken")
			setSessions(w, r)
			http.Redirect(w, r, "/signup", 302)
		}
	} else {
		// we have an incorrect email address, so we flash a message and redirect back to the signup
		session30m.AddFlash("Error: invalid email address")
		setSessions(w, r)
		http.Redirect(w, r, "/signup", 302)
	}
}
func changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	log.Notice("/change-password")
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


	err := allTemplates.ExecuteTemplate(w, "change-password", context)
	if_err_panic("error executing template: ", err)
}

func authInternalChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	// this is an internal handler that does not generate a page
	// it performs action altering some store values, then redirects

	log.Notice("/auth/internal/change-password")
	// sessions & user management ///////////////////////////
	session30m, _ := getSessions(r)

	var userState UserState
	var changePasswordCredentials ChangePasswordCredentials

	// type assert userState and extendedState
	// or keep the ones you got
	var ok bool
	if userState, ok = session30m.Values["userState"].(UserState); ok {
		// RTA ok
	} else {
		// RTA faild, default value is ok
	}

	// parse the forms
	err := r.ParseForm()
	if_err_panic("78ej53: Parse form failed", err)

	// create the gorilla/scheme decoder
	decoder := schema.NewDecoder()
	decoder.ZeroEmpty(true)

	// get email and password from form
	err = decoder.Decode(&changePasswordCredentials, r.PostForm)
	if_err_panic("67hj529: Deconding failed", err)

	// get user record with that email of fail
	// scan a User record to check if it exists and if it's active
	// prepare the statement
	stmt, err := db.Prepare("SELECT user_id, password FROM users WHERE email = ? LIMIT 1")
	if_err_panic("45jn61: malformed statement", err)
	defer stmt.Close()

	// perform the query
	rows, err := stmt.Query(userState.Email)
	if_err_panic("14Kfg: failed statement", err)

	// create a temp UserRow
	var scanUser UserRow

	for rows.Next() {
		// scan record inside scanUser's fields
		err := rows.Scan(&scanUser.UserID, &scanUser.Password)
		scanUser.Enabled = true
		if_err_panic("34DhV: error during scan", err)
	}
	defer rows.Close()

	// first check if supplied passwords match
	// no need to go on otherwise
	if changePasswordCredentials.NewPassword == changePasswordCredentials.NewPasswordConfirmation {
		// get the encrypted password and check if match
		err = bcrypt.CompareHashAndPassword([]byte(scanUser.Password), []byte(changePasswordCredentials.OldPassword))
		if err == nil {
			// match SUCCESS
			// nil error means password match
			// proceed to  write the new encrypted password to the db store
			// also write a confirmation flash message
			stmt, err := db.Prepare("UPDATE users SET password = ? WHERE user_id = ?")
			newPassword, err := bcrypt.GenerateFromPassword([]byte(changePasswordCredentials.NewPassword), 8)
			if_err_panic("df37ui: bcrypt encription error", err)
			log.Debug("New password is= %#+v", newPassword)
			// update record
			stmt.Exec(newPassword, scanUser.UserID)
			if_err_panic("57ju61: malformed statement", err)
			defer stmt.Close()

			if_err_panic("14Kfg: failed statement", err)
			session30m.AddFlash("Success: password succesfully updated")
			setSessions(w, r)
			http.Redirect(w, r, "/", 302)
		} else {
			// match FAIL
			session30m.AddFlash("Fail: incorrect current password")
			setSessions(w, r)
			http.Redirect(w, r, "/change-password", 302)
		}

	} else {
		// new password and confirmation does not match
		session30m.AddFlash("Fail: new password and confirmation does not match")
		setSessions(w, r)
		http.Redirect(w, r, "/change-password", 302)
	}
}

func authInternalRequestResetHandler(w http.ResponseWriter, r *http.Request) {
	// this is an internal handler that does not generate a page
	// it performs action altering some store values, then redirects

	log.Notice("/auth/internal/request-reset/{recoveryCode}")
	session30m, _ := getSessions(r)

	// parse the forms
	err := r.ParseForm()
	if_err_panic("33ef53: Parse form failed", err)

	// create the gorilla/scheme decoder
	decoder := schema.NewDecoder()
	decoder.ZeroEmpty(true)

	var resetRequestCredentials ResetRequestCredentials

	// decode signup struct from form
	err = decoder.Decode(&resetRequestCredentials, r.PostForm)
	if_err_panic("47ej66: Decoding failed", err)

	log.Debug("email= %+#v", resetRequestCredentials.Email)

	// prepare a statement to find one row with the correct activation code

	stmt, err := db.Prepare("SELECT user_id, nick FROM users WHERE email = ? LIMIT 1;")
	if_err_panic("35H77: malformed statement", err)
	defer stmt.Close()

	// execute the query and hope to get a result of one row
	rows, err := stmt.Query(resetRequestCredentials.Email)

	defer rows.Close()

	// result is good, rows.Next() returns a bool = true if there's a row to read
	if rows.Next() {

		// we generate and write the recovery code
		// prepare the statement
		var userID int64
		var nick string
		var recoveryCode int64
		var mailBody string

		rows.Scan(&userID, &nick)
		recoveryCode = minimapID()

		stmt, err = db.Prepare("UPDATE users SET recovery_code = ? WHERE user_id = ?")
		if_err_panic("35H67: malformed statement", err)
		defer stmt.Close()

		// perform the query
		_, err = stmt.Exec(recoveryCode, userID)
		if_err_panic("17KCC: failed exec", err)

		// send the email
		sg := sendgrid.NewSendGridClient("ilnerdchuck", "Numero98")
		message := sendgrid.NewMail()
		message.AddTo(resetRequestCredentials.Email)
		message.AddToName(nick)
		message.SetSubject("Password reset request for " + resetRequestCredentials.Email)

		mailBody  = "Hi, we just got a reset request for your account. "
		mailBody += "Please click the following link to start the reset procedure."
		mailBody += "http://cm.avero.it:33000/reset-password/"
		mailBody += strconv.FormatInt(recoveryCode, 10)

		message.SetText(mailBody)
		message.SetFrom("do_not_reply@minimap.avero.com")
		if r := sg.Send(message); r == nil {
			fmt.Println("Email with recovery code sent!")
		} else {
			fmt.Println(r)
		}

		// user record succesfully created
		session30m.AddFlash("Success: recovery email sent. Please follow instructions on email.")
	} else {
		// we got an email not pertaining to a user, so we flash a message and redirect back to login
		session30m.AddFlash("Error: invalid email address")
	}

	setSessions(w, r)
	http.Redirect(w, r, "/login", 302)
}

func authInternalResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	// this is an internal handler that does not generate a page
	// it performs action altering some store values, then redirects

	log.Notice("/auth/internal/reset-password/{recoveryCode}")
	// sessions & user management ///////////////////////////
	session30m, _ := getSessions(r)

	var resetPasswordCredentials ResetPasswordCredentials

	// parse the forms
	err := r.ParseForm()
	if_err_panic("77ek53: Parse form failed", err)

	// create the gorilla/scheme decoder
	decoder := schema.NewDecoder()
	decoder.ZeroEmpty(true)

	// get email and password from form
	err = decoder.Decode(&resetPasswordCredentials, r.PostForm)
	if_err_panic("67hj51: Decoding failed", err)

	// and get recoverypath from route
	vars := mux.Vars(r)
	resetPasswordCredentials.RecoveryCode = vars["recoveryCode"]

	// get user record with that email of fail
	// scan a User record to check if it exists and if it's active
	// prepare the statement
	stmt, err := db.Prepare("SELECT user_id FROM users WHERE recovery_code = ? LIMIT 1")
	if_err_panic("45jn69: malformed statement", err)
	defer stmt.Close()

	// perform the query
	rows, err := stmt.Query(resetPasswordCredentials.RecoveryCode)
	if_err_panic("14Kfk: failed statement", err)

	// create a temp UserRow
	var scanUser UserRow

	for rows.Next() {
		// scan record inside scanUser's fields
		err := rows.Scan(&scanUser.UserID)
		scanUser.Enabled = true
		if_err_panic("34DhV: error during scan", err)
	}
	defer rows.Close()

	// first check if supplied passwords match
	// no need to go on otherwise
	if resetPasswordCredentials.NewPassword == resetPasswordCredentials.NewPasswordConfirmation {
		// nil error means password match
		// proceed to  write the new encrypted password to the db store
		// also write a confirmation flash message
		stmt, err := db.Prepare("UPDATE users SET password = ?, recovery_code = 0 WHERE user_id = ?")
		newPassword, err := bcrypt.GenerateFromPassword([]byte(resetPasswordCredentials.NewPassword), 8)
		if_err_panic("df39uk: bcrypt encription error", err)
		log.Debug("New password is= %#+v", newPassword)
		// update record
		stmt.Exec(newPassword, scanUser.UserID)
		if_err_panic("57ju66: malformed statement", err)
		defer stmt.Close()

		if_err_panic("13Kfp: failed statement", err)
		session30m.AddFlash("Success: password succesfully updated")
		setSessions(w, r)
		http.Redirect(w, r, "/login", 302)

	} else {
		// new password and confirmation does not match
		session30m.AddFlash("Fail: new password and confirmation does not match")
		setSessions(w, r)
		http.Redirect(w, r, r.RequestURI, 302)
	}
}

func authInternalActivateHandler(w http.ResponseWriter, r *http.Request) {
	// this is an internal handler that does not generate a page
	// it performs action altering some store values, then redirects

	log.Notice("/auth/internal/activate/{activationCode}")
	session30m, _ := getSessions(r)

	vars := mux.Vars(r)
	activationCode := vars["activationCode"]

	log.Debug("activationCode= %+#v", activationCode)

	// prepare a statement to find one row with the correct activation code

	stmt, err := db.Prepare("SELECT user_id FROM users WHERE activation_code = ?;")
	if_err_panic("35H66: malformed statement", err)
	defer stmt.Close()

	// execute the query and hope to get a result of one row
	rows, err := stmt.Query(activationCode)

	defer rows.Close()

	// result is good, rows.Next() returns a bool = true if there's a row to read
	if rows.Next() {

		// we remove the activatiod code
		// prepare the statement

		stmt, err = db.Prepare("UPDATE users SET activation_code = 0 WHERE activation_code = ?")
		if_err_panic("35H67: malformed statement", err)
		defer stmt.Close()

		// perform the query
		_, err = stmt.Exec(activationCode)
		if_err_panic("17KBB: failed exec", err)

		// user record succesfully created
		session30m.AddFlash("Success: user correctly activated. Please log in.")
	} else {
		// we have an incorrect activation code, so we flash a message and redirect back to login
		session30m.AddFlash("Error: invalid email address")
	}

	setSessions(w, r)
	http.Redirect(w, r, "/login", 302)
}

func requestResetHandler(w http.ResponseWriter, r *http.Request) {
	log.Notice("/request-reset")
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


	err := allTemplates.ExecuteTemplate(w, "request-reset", context)
	if_err_panic("error executing template: ", err)
}

func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	log.Notice("/reset-password")
	// sessions & user management ///////////////////////////
	session30m, session1y := getSessions(r)
	userState, extendedState := assertUser(w, r, session30m, session1y)

	// we need to propagate the recovery code in the url trough the correct auth handler
	// so we get it here and pass it to the template renderer in a specific context


	vars := mux.Vars(r)
	recoveryCode := vars["recoveryCode"]

	// context & templates ///////////////////////////////////
	var contextFlashes []string

	for _, flash := range session30m.Flashes() {
		if str, ok := flash.(string); ok {
			contextFlashes = append(contextFlashes, str)
		}
	}


	// call this BEFORE writing to the http.ResponseWriter
	setSessions(w, r)

	context := Context{r, userState, extendedState, contextFlashes}
	resetPasswordContext := ResetPasswordContext{context, recoveryCode}

	if r.URL.Query().Get("debug") == "true" {
		fmt.Fprintf(w, "--- CONTEXT---\n")
		fmt.Fprintf(w, "%#+v\n", resetPasswordContext)
		fmt.Fprintf(w, "--- GENERATED PAGE ---\n")
	}

	err := allTemplates.ExecuteTemplate(w, "reset-password", resetPasswordContext)
	if_err_panic("error executing template: ", err)
}

func authInternalLogoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Notice("/auth/internal/logout")
	session30m, session1y := getSessions(r)
	expireSessions(w, r, session30m, session1y)
	// call this BEFORE writing to the http.ResponseWriter
	setSessions(w, r)
	http.Redirect(w, r, "/", 302)
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
	stmt, err := db.Prepare("SELECT user_id, nick, email, password FROM users WHERE activation_code = 0 AND email = ? LIMIT 1")
	if_err_panic("35H67: malformed statement", err)
	defer stmt.Close()

	// perform the query
	rows, err := stmt.Query(loginCredentials.Email)
	if_err_panic("17KBB: failed statement", err)

	// create a temp UserRow
	var scanUser UserRow

	for rows.Next() {
		// scan record inside scanUser's fields
		err := rows.Scan(&scanUser.UserID, &scanUser.Nick, &scanUser.Email, &scanUser.Password)
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

func testHandler(w http.ResponseWriter, r *http.Request) {
log.Notice("/test")
session30m, session1y := getSessions(r)
userState, extendedState := assertUser(w, r, session30m, session1y)

log.Debug("userState is: %#+v", userState)
log.Debug("extendedState is: %#+v", extendedState)

// route specific work
//	session30m.AddFlash("flash1")
//	session30m.AddFlash("flash2")
//	session30m.AddFlash("flash3")

/*
	// Add a value
	// session30m.Values["try"] = nil

	if str, ok := session30m.Values["try"].(string); ok {
		fmt.Printf("Returned:" + str + "\n")
	} else {
		fmt.Printf("Type assert failed(session30m)\n")
	}
*/

// Delete session.
//	session.Options.MaxAge = -1

/*
	// prepare a statement
	stmt, err := db.Prepare("SELECT user_id, email FROM users WHERE user_id > ?")
	if_err_panic("57HNB: malformed statement", err)
	defer stmt.Close()

	rows, err := stmt.Query(3487892730368)
	if_err_panic("12SDC: failed statement", err)

	var Users []User
	var scanuser User

	for rows.Next() {
		err := rows.Scan(&scanuser.Id, &scanuser.Name)
		if_err_panic("23DFV: error during scan", err)
		//fmt.Println(id, name)
		Users = append(Users, scanuser)
	}

	rows.Close()
*/

var contextFlashes []string

for _, flash := range session30m.Flashes() {
if str, ok := flash.(string); ok {
contextFlashes = append(contextFlashes, str)
}
}

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
id:=minimapID()
	log.Debug("id is: %#+v", id)

err := allTemplates.ExecuteTemplate(w, "test", context)
if_err_panic("error executing template: ", err)
}
