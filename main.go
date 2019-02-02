package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func main() {
	log.Println("[MAIN] Starting...")
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
	router.HandleFunc("/addone", addone)
	router.HandleFunc("/echo", echo)
	router.HandleFunc("/v1", v1)

	http.ListenAndServe(":"+port, router)
}
