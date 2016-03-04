package main

import (
	"flag"
	"fmt"
	"html/template"
	"path/filepath"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"database/sql"
	"gopkg.in/boj/redistore.v1"
)

// PACKAGE GLOBALS//////////////////////////////////////////////////////////////////////////////////////////////////////

var log = logging.MustGetLogger("example")
var store *redistore.RediStore
var port int
var db *sql.DB
var allTemplates = template.New("home")
var templates []string

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

	const defaultPort = 33000

	flag.IntVar(&port, "port", 0, "sets the listen port on localhost")
	flag.Parse()

	if port == 0 {
		port = defaultPort
		log.Warning("No port supplied with --port, falling back to default port: %v", defaultPort)
	}

	r := mux.NewRouter()

	r.HandleFunc("/list", listHandler)

	http.Handle("/", r)

	log.Notice("Listening on port: %v", port)

	err := http.ListenAndServe(fmt.Sprint(":", port), r)
	if err != nil {
		panic(err)
	}
}

