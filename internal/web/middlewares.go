package web

import (
	"crypto/sha256"
	"crypto/subtle"
	"log"
	"net/http"
)

type BasicAuth struct {
	http.Handler
	username       string
	passwordSha224 [28]byte
}

func unauthorizedHandler(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Basic")
	w.WriteHeader(http.StatusUnauthorized)
}

func NewBasicAuth(handler http.Handler, username, password string) *BasicAuth {
	return &BasicAuth{
		Handler: handler, username: username, passwordSha224: sha256.Sum224([]byte(password)),
	}
}

func (ba *BasicAuth) verify(inputUsername, inputPassword string) bool {
	if inputUsername != ba.username {
		return false
	}
	inputPasswordHash := sha256.Sum224([]byte(inputPassword))
	return subtle.ConstantTimeCompare(ba.passwordSha224[:], inputPasswordHash[:]) == 1
}

func (ba *BasicAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	inputUsername, inputPassword, ok := r.BasicAuth()
	if !ok {
		unauthorizedHandler(w)
	} else if !ba.verify(inputUsername, inputPassword) {
		unauthorizedHandler(w)
		log.Printf("[web] %s - HTTP Basic Auth Failed: %s", r.RemoteAddr, inputUsername)
	} else {
		ba.Handler.ServeHTTP(w, r)
	}
}
