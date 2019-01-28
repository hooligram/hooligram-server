package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
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

	mysqlUsername := os.Getenv("MYSQL_USERNAME")
	if mysqlUsername == "" {
		log.Fatal("MYSQL_USERNAME must be set. Exiting...")
	}

	mysqlPassword := os.Getenv("MYSQL_PASSWORD")
	if mysqlPassword == "" {
		log.Fatal("MYSQL_PASSWORD must be set. Exiting...")
	}

	mysqlDbName := os.Getenv("MYSQL_DB_NAME")
	if mysqlDbName == "" {
		log.Fatal("MYSQL_DB_NAME must be set. Exiting...")
	}

	db, err := sql.Open("mysql", mysqlUsername+":"+mysqlPassword+"@/"+mysqlDbName)
	if err != nil {
		log.Fatal("Error setting up MySQL DB connection. Exiting...")
	}

	db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Can't connect to MySQL DB. Exiting...")
	}

	router := mux.NewRouter()
	router.HandleFunc("/addone", addone)
	router.HandleFunc("/echo", echo)
	router.HandleFunc("/v1", v1)
	http.ListenAndServe(":"+port, router)
}
