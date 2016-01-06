package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/alexgear/checker/common"
	"github.com/alexgear/checker/config"
	"github.com/alexgear/checker/datastore"
	"github.com/gorilla/mux"
)

var err error

func postDataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	vars := mux.Vars(r)
	r.ParseForm()
	//log.Printf("postDataHandler: %#v", vars, r.Form)
	response := common.Response{}
	response.Latency, err = time.ParseDuration(r.Form["latency"][0])
	if err != nil {
		log.Println("Failed to parse duration:", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response.Time, err = time.Parse(time.RFC3339, r.Form["time"][0])
	if err != nil {
		log.Println("Failed to parse time:", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response.Status, err = strconv.ParseBool(r.Form["status"][0])
	if err != nil {
		log.Println("Failed to parse status:", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = datastore.Write(vars["ief"], response)
	if err != nil {
		log.Println("Failed to write to db:", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//response := SMSResponse{Text: sms.Body, UUID: sms.UUID, Status: sms.Status}
	//toWrite, err := json.Marshal(response)
	//if err != nil {
	//	log.Println(err)
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//w.Write(toWrite)
	return
}

func InitServer() error {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/v1/{ief}", postDataHandler).Methods("POST")
	bind := fmt.Sprintf("%s:%d", config.C.ListenHost, config.C.ListenPort)
	log.Println("listening on: ", bind)
	return http.ListenAndServe(bind, router)
}
