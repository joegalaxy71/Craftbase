package main
import(
	"os"
	"net/http"
	"github.com/gorilla/sessions"
)
func getSessions(r *http.Request) (*sessions.Session, *sessions.Session) {

	// Get or create sessions (both short and long lived)
	session30m, err := store.Get(r, "session30m")
	if_err_panic("23gh34: error creating session with 30m expiry", err)
	session30m.Options.MaxAge = 60 * 30
	session30m.Options.HttpOnly = true

	session1y, err := store.Get(r, "session1y")
	if_err_panic("12gh34: error creating session with 1y expiry", err)
	session1y.Options.MaxAge = 3600 * 24 * 365
	session1y.Options.HttpOnly = true

	return session30m, session1y
}

func setSessions(w http.ResponseWriter, r *http.Request) {
	// save sessions, thus generating corresponding cookies
	err := sessions.Save(r, w)
	if_err_panic("12ds34: error saving session", err)

}
func ReLogin(userState *UserState) {
	// scan a User record to check if it exists and if it's active
	// prepare the statement
	stmt, err := db.Prepare("SELECT user_id, nick, email, password, lang, lat, lng, rng, image, status, message, ftsync FROM users WHERE enabled = true AND user_id = ? LIMIT 1")
	if_err_panic("57HNB: malformed statement", err)
	defer stmt.Close()

	// perform the query
	rows, err := stmt.Query(userState.UserID)
	if_err_panic("17KLC: failed statement", err)

	// create a temp UserRow
	var scanUser UserRow

	for rows.Next() {
		// scan record inside scanUser's fields
		err := rows.Scan(&scanUser.UserID, &scanUser.Nick, &scanUser.Email, &scanUser.Password, &scanUser.Lang)
		scanUser.Enabled = true
		if_err_panic("45123: error during scan", err)
	}

	rows.Close()

	// if we correctly fetched a record we update the userState
	if scanUser.Enabled == true {
		userState.Logged = true
		userState.UserID = scanUser.UserID
		userState.Nick = scanUser.Nick
		userState.Email = scanUser.Email
		userState.Lang = scanUser.Lang

	}

	// if we don't find any record, the userState stays
	// as it is, unlogged and with zeroed UserID
}

func assertUser(w http.ResponseWriter, r *http.Request, session30m, session1y *sessions.Session) (UserState, ExtendedState) {
	// declare/define/init vars
	var userState UserState
	var extendedState ExtendedState
	var ok bool

	// userState is mandatory on any route

	// retrieve a UserState from session30m and type assert it
	if userState, ok = session30m.Values["userState"].(UserState); ok {
		// the user is either INTERACTING OR WANDERING, still need to RTA extendedState
		// to have updated status
		extendedState, ok = session1y.Values["extendedState"].(ExtendedState)
		if userState.Logged {
			// user is INTERACTING and we persist some fields in extendedState
			log.Debug("user is INTERACTING")
			if userState.RememberMe == true {
				extendedState.UserID = userState.UserID
			}
			extendedState.Lang = userState.Lang
		} else {
			// user is WANDERING and we persist a smaller subset and zero UserID
			log.Debug("user is WANDERING")
			extendedState.UserID = 0
		}
	} else {
		// session30m userState RTA failed
		// user may be RETURN-INTERACTING, RETURN-WANDERING or REALLYNEW
		log.Debug("Type assert failed(30m): no userState in session\n")
		// create an empty WANDERING userstate
		userState = UserState{}
		log.Debug("Created empty UserState\n")
		// RTA extendedState to determine if the user is RETURN-INTERACTING
		// or RETURN-WANDERING .... if it's missing its REALLYNEW
		if extendedState, ok = session1y.Values["extendedState"].(ExtendedState); ok {
			// by checking the UserID we can discern RETURN-INTERACTING or RETURN-WANDERING
			if extendedState.UserID != 0 {
				// user is RETURN-INTERACTING
				log.Debug("user is RETURN-INTERACTING")
				// copy relevant data back to userState
				userState.UserID = extendedState.UserID
				// ReLogin will fetch a record from the user table,
				// check is the user is enabled and set userState.Logged
				ReLogin(&userState)
			} else {
				// user is RETURN-WANDERING, we copy only a small info subset
				log.Debug("user is RETURN-WANDERING")
				userState.Lang = extendedState.Lang
			}
		} else {
			// user is REALLYNEW
			log.Debug("user is REALLYNEW")
			log.Debug("Type assert failed(1ym): no extendedState in session1y\n")
			// determine language
			lang := determineLang(r, "")
			log.Debug("determined lang: %v", lang)
			// create an empty extendedState
			extendedState = ExtendedState{0, lang, ""}
			log.Debug("Created empty extendedState\n")
		}
	}

	// we rewrite both userState and extendedState
	// thus extending duration by session duration (30m and 1y)
	session30m.Values["userState"] = userState
	session1y.Values["extendedState"] = extendedState

	return userState, extendedState
}

func cleanup(c chan os.Signal) {
	<-c
	log.Warning("Got os.Interrupt: cleaning up")

	// exiting gracefully
	os.Exit(0)
}

func if_err_panic (msg string, e error) {
	if e != nil {
		log.Error(msg, e)
		panic(e)
	}
}
