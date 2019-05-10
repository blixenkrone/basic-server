package main

import (
	"crypto/tls"
	"flag"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/acme/autocert"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
)

var (
	host = flag.String("host", "", "What host are you using?")
)

func main() {
	flag.Parse()
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from go-pro!"))
	})

	cm := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(*host),
		Cache:      autocert.DirCache("/certs"),
	}

	httpsSrv := &http.Server{
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
		Addr:              ":https",
		TLSConfig: &tls.Config{
			PreferServerCipherSuites: true,
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
		},
		Handler: cm.HTTPHandler(r),
	}

	if *host == "" {
		httpsSrv.Addr = "localhost:8085"
		log.Printf("Serving on addr: %s", httpsSrv.Addr)
		log.Fatal(httpsSrv.ListenAndServeTLS("certs/insecure_cert.pem", "certs/insecure_key.pem"))
	}

	err := useHTTP2(httpsSrv)
	if err != nil {
		log.Fatal(err)
	}
	httpsSrv.TLSConfig.GetCertificate = cm.GetCertificate
	log.Info("Serving on port 443, authenticating for https://", *host)
	log.Fatal(httpsSrv.ListenAndServeTLS("", ""))
}

func useHTTP2(httpsSrv *http.Server) error {
	http2Srv := http2.Server{}
	err := http2.ConfigureServer(httpsSrv, &http2Srv)
	if err != nil {
		return err
	}
	return nil
}

func lookupEnv(val, fallback string) string {
	v, ok := os.LookupEnv(val)
	if !ok {
		log.Printf("Error getting env, using: %s", fallback)
		return fallback
	}
	return v
}
