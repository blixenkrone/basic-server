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
	env          = lookupEnv("ENV", "development")
	insecurePort = flag.String("insecure-bind", ":80", "host/port to bind on for insecure (HTTP) traffic")
	securePort   = flag.String("secure-bind", ":443", "host/port to bind on for secure (HTTPS) traffic")
	sitePort     = flag.String("site-port", "3000", "port to http forward")
	// siteDomain   = flag.String("site-domain", "git.xeserv.us", "site port")
)

func main() {

	startServer()
}

func startServer() {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello!"))
	})
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
			// GetCertificate: certManager().GetCertificate,
		},
		Handler: r,
	}
	err := useHTTP2(httpsSrv)
	if err != nil {
		log.Fatal(err)
	}

	if env == "development" {
		httpsSrv.Addr = "localhost:8085"
		log.Printf("Serving on addr: %s", httpsSrv.Addr)
		log.Fatal(httpsSrv.ListenAndServeTLS("certs/insecure_cert.pem", "certs/insecure_key.pem"))
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

func useHTTP2(httpsSrv *http.Server) error {
	http2Srv := http2.Server{}
	err := http2.ConfigureServer(httpsSrv, &http2Srv)
	if err != nil {
		return err
	}
	return nil
}

func certManager() *autocert.Manager {
	mng := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("https://localhost:8085"),
		// Cache:      autocert.DirCache("./"),
	}
	return &mng
}
