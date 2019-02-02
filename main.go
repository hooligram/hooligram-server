package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func main() {
	defer db.Close()

	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("[MAIN] PORT must be set. Exiting...")
	}

	twilioAPIKey = os.Getenv("TWILIO_API_KEY")

	if twilioAPIKey == "" {
		log.Fatal("[MAIN] TWILIO_API_KEY must be set. Exiting...")
	}

	router := mux.NewRouter()
	router.HandleFunc("/v1", v1)

	http.ListenAndServe(":"+port, router)
}
