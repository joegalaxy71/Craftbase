package main

import (
	"database/sql"
	"flag"
	"fmt"
	"math/rand"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"gopkg.in/boj/redistore.v1"
	"html/template"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
	"os/signal"
	"path/filepath"
)

// PACKAGE GLOBALS//////////////////////////////////////////////////////////////////////////////////////////////////////

var log = logging.MustGetLogger("example")
var store *redistore.RediStore
var port int
var db *sql.DB
var allTemplates = template.New("home")
var templates []string
var minirand *rand.Rand

// INIT ////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	var err error



	// logging
	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)
	format := logging.MustStringFormatter(
		"%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}")
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")
	logging.SetBackend(backend1Leveled, backend2Formatter)

	db, err = sql.Open("mysql", "root:Numero98@tcp(localhost:3306)/craftbase")
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	// redistore init
	store, err = redistore.NewRediStore(10, "tcp", ":6379", "", []byte("secret-key"))
	if_err_panic("12gh34: error creating store", err)

	basepath := "/var/go/craftbase/src/craftbase/templates/"

	filepath.Walk(basepath, addTemplate)

	randsource := rand.NewSource(time.Now().UnixNano())
	minirand = rand.New(randsource)

	//fp := path.Join(basepath, "*.tmpl")

	allTemplates, err = template.ParseFiles(templates...)
	checkErr("error parsing template: ", err)

	log.Info("routes.go: init completed\n")

	log.Info("main.go: init completed\n")
}

// MAIN ////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func main() {
	defer db.Close()

	// handle ^c (os.Interrupt)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go cleanup(c)

	const (defaultPort = 33000
			sendgridAPIKEY = ""
	)
	flag.IntVar(&port, "port", 0, "sets the listen port on localhost")
	flag.Parse()

	if port == 0 {
		port = defaultPort
		log.Warning("No port supplied with --port, falling back to default port: %v", defaultPort)
	}

	router := mux.NewRouter()

	//pages
	router.HandleFunc("/", homeHandler)
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/signup", signUpHandler)
	router.HandleFunc("/change-password", changePasswordHandler)
	router.HandleFunc("/request-reset", requestResetHandler)
	router.HandleFunc("/reset-password/{recoveryCode}", resetPasswordHandler)
	router.HandleFunc("/test", testHandler)
	//
	////auth
	//
	router.HandleFunc("/auth/internal/signup", authInternalSignUpHandler)
	router.HandleFunc("/auth/internal/activate/{activationCode}", authInternalActivateHandler)
	router.HandleFunc("/auth/internal/login", authInternalLoginHandler)
	router.HandleFunc("/auth/internal/change-password", authInternalChangePasswordHandler)
	router.HandleFunc("/auth/internal/request-reset", authInternalRequestResetHandler)
	router.HandleFunc("/auth/internal/reset-password/{recoveryCode}", authInternalResetPasswordHandler)
	router.HandleFunc("/auth/internal/logout", authInternalLogoutHandler)

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("/var/go/craftbase/bin/")))

	http.Handle("/", router)

	log.Notice("Listening on port: %v", port)

	err := http.ListenAndServe(fmt.Sprint(":", port), router)
	if err != nil {
		panic(err)
	}
}
