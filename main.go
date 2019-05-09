package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World! I'm a Gopher!"))
	})

	log.Info("Running on port 8085")

	if err := http.ListenAndServe(":8085", r); err != nil {
		log.Fatal(err)
	}
}
