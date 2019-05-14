package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/acme/autocert"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
)

var (
	host       = flag.String("host", "", "What host are you using?")
	production = flag.Bool("production", false, "Is it production?")
)

var jwtKey = []byte("thiskeyiswhat")

// JWTClaims -
type JWTClaims struct {
	Username string `json:"username"`
	Claims   jwt.StandardClaims
}

// JWTCreds for at user to get JWT
type JWTCreds struct {
	Username string `json:"username"`
}

func generateJWT(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var creds JWTCreds
		if err := decodeJSON(r.Body, creds); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		expirationTime := time.Now().Add(5 * time.Minute).Unix()
		claims := &JWTClaims{
			Username: creds.Username,
			Claims: jwt.StandardClaims{
				ExpiresAt: expirationTime,
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims.Claims)
		signedToken, err := token.SignedString(jwtKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if err := json.NewEncoder(w).Encode(signedToken); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// isJWTAuth requires routes to possess a JWToken
func isJWTAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("token")
		// ! If the JWT failed
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, http.ErrNoCookie.Error(), http.StatusUnauthorized)
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		// * If JWT succeeded
		token, err := jwt.Parse(c.Value, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				http.Error(w, http.ErrMissingContentLength.Error(), 400)
			}
			return jwtKey, nil
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if token.Valid {
			log.Info("Valid JWT")
			next(w, r)
		}
	})
}

func secureMessage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Secret hello from go-pro!"))
}

func main() {
	http.HandleFunc("/secure", isJWTAuth(secureMessage))
	http.HandleFunc("/authenticate", generateJWT)

	// Create auto-certificate https server
	m := autocert.Manager{
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
		Handler: nil,
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

	if err := useHTTP2(httpsSrv); err != nil {
		log.Warnf("Error with HTTP2 %s", err)
	}

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
		if fallback != "" {
			log.Printf("Error getting env, using: %s", fallback)
			return fallback
		}
		log.Error("Error getting fallback")
	}
	return v
}

func decodeJSON(r io.Reader, val JWTCreds) error {
	err := json.NewDecoder(r).Decode(&val)
	if err != nil {
		return err
	}
	return nil
}
