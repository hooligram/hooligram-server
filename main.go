package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/hooligram/hooligram-server/delivery"
	"github.com/hooligram/hooligram-server/v2"
)

const mainTag = "main"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal(fmt.Sprintf("[%v] PORT not set", mainTag))
	}

	router := mux.NewRouter()
	router.HandleFunc("/v2", v2.V2)

	go delivery.DeliverMessage()

	http.ListenAndServe(":"+port, router)
}
