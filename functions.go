package main
import(
	"os"

)

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
