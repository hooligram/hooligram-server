package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("PORT must be set. Exiting...")
	}

	twilioAPIKey = os.Getenv("TWILIO_API_KEY")

	if twilioAPIKey == "" {
		log.Fatal("TWILIO_API_KEY must be set. Exiting...")
	}

	router := mux.NewRouter()
	router.HandleFunc("/addone", addone)
	router.HandleFunc("/echo", echo)
	router.HandleFunc("/v1", v1)
	http.ListenAndServe(":"+port, router)
}
