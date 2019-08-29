package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type WithCORS struct {
	r *mux.Router
}

var firebaseApp = initializeFirebaseApp()

func main() {
	configuration, _ := Configuration()

	log.Println("Starting server on 0.0.0.0:" + configuration.Port)
	log.Fatal(http.ListenAndServe(":"+configuration.Port, &WithCORS{NewRouter()}))
}

// Simple wrapper to Allow CORS
func (s *WithCORS) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	origin := req.Header.Get("Origin")
	res.Header().Set("Access-Control-Allow-Origin", origin)
	res.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	res.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Origin")
	res.Header().Set("Access-Control-Expose-Headers", "Content-Range")

	// Stop here for a Preflighted OPTIONS request.
	if req.Method == "OPTIONS" {
		return
	}
	// Lets Gorilla work
	s.r.ServeHTTP(res, req)
}
