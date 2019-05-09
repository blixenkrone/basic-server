package main

import (
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

func main() {
	val := lookupEnv("ENV", "development")
	log.Info(val)
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p := mux.Vars(r)
		for key, val := range p {
			w.Write([]byte("Hello World! I'm a Gopher! Param: %s", p[key]))
		}
	})

	log.Info("Running on port 8085")

	if err := http.ListenAndServe(":8085", r); err != nil {
		log.Fatal(err)
	}
}

func lookupEnv(val, fallback string) string {
	v, ok := os.LookupEnv(val)
	if !ok {
		log.Printf("Error getting env, using: %s", fallback)
		return fallback
	}
	return v
}
