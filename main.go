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
	host       = flag.String("host", "", "What host are you using?")
	production = flag.Bool("production", false, "Is it production?")
)

func main() {
	flag.Parse()
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from go-pro!"))
	})

	// Create auto-certificate https server
	m := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(*host),
		Cache:      autocert.DirCache("/certs"),
	}

	httpsSrv := &http.Server{
		// ReadTimeout:       5 * time.Second,
		// WriteTimeout:      10 * time.Second,
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
		Handler: r,
	}

	// Serve on localhost with localhost certs if no host provided
	if *host == "" {
		httpsSrv.Addr = "localhost:8085"
		log.Info("Serving on http://localhost:8085")
		// log.Fatal(httpsSrv.ListenAndServeTLS("./certs/insecure_cert.pem", "./certs/insecure_key.pem"))
		log.Fatal(httpsSrv.ListenAndServe())
	}

	// Create server for redirecting HTTP to HTTPS
	httpSrv := &http.Server{
		Addr:         ":http",
		ReadTimeout:  httpsSrv.ReadTimeout,
		WriteTimeout: httpsSrv.WriteTimeout,
		IdleTimeout:  httpsSrv.IdleTimeout,
		Handler:      m.HTTPHandler(nil),
	}

	// if err := useHTTP2(httpsSrv); err != nil {
	// 	log.Warnf("Error with HTTP2 %s", err)
	// }

	go func() {
		log.Fatal(httpSrv.ListenAndServe())
	}()

	httpsSrv.TLSConfig.GetCertificate = m.GetCertificate
	log.Info("Serving on https, authenticating for https://", *host)
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
