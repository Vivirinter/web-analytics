package main

import (
	"log"
	"net/http"

	"./controllers"
	"./utils"
	"github.com/gorilla/mux"
)

func main() {
	utils.RC = utils.GetRedisConnection()
	defer utils.RC.Close()
	r := mux.NewRouter()
	r.HandleFunc("/", controllers.SetPageview).Methods("POST")
	r.HandleFunc("/pageviews", controllers.GetPageviews).Methods("GET")
	r.HandleFunc("/uniques", controllers.GetUniques).Methods("GET")

	log.Println("(Web Analytics v1.0) : Listening on localhost:8080 for requests...")
	log.Fatal(http.ListenAndServe("localhost:8080", r))
}
