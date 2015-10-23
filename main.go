package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/pprof"
	"database/sql"
)

// PACKAGE GLOBALS//////////////////////////////////////////////////////////////////////////////////////////////////////

var log = logging.MustGetLogger("example")

// Example format string. Everything except the message has a custom color
// which is dependent on the log level. Many fields have a custom output
// formatting too, eg. the time returns the hour down to the milli second.
var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}")

var port int
var f *os.File
var db sql.DB

// INIT ////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func init() {
	var err error

	// logging
	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")
	logging.SetBackend(backend1Leveled, backend2Formatter)

	//create file for profiling
	f, err = os.Create("ubiquy.cpuprofile")
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("mysql", "root:Numero98@/craftbase")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	log.Info("main.go: init completed\n")
}

// MAIN ////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func main() {

	// handle ^c (os.Interrupt)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go cleanup(c)

	// start CPU profiling
	pprof.StartCPUProfile(f)

	const defaultPort = 3333

	flag.IntVar(&port, "port", 0, "sets the listen port on localhost")
	flag.Parse()

	if port == 0 {
		port = defaultPort
		log.Warning("No port supplied with --port, falling back to default port: %v", defaultPort)
	}

	r := mux.NewRouter()

	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/products", productsHandler)
	r.HandleFunc("/articles", articlesHandler)

	http.Handle("/", r)

	log.Notice("Listening on port: %v", port)

	err := http.ListenAndServe(fmt.Sprint(":", port), r)
	if err != nil {
		panic(err)
	}
}

// FUNCTIONS ////////////////////////////////////////////////////////////////////////////////////////////////////////

func cleanup(c chan os.Signal) {
	<-c
	log.Warning("Got os.Interrupt: cleaning up")

	// stopping profiling and closing file
	pprof.StopCPUProfile()
	f.Close()

	// exiting gracefully
	os.Exit(0)
}
