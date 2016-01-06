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
	response.Time, err = time.Parse(time.RFC3339Nano, r.Form["time"][0])
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
	return
}

// getGraphHandler writes a self-contained HTML page with an interactive plot
// of the latencies from datastore, built with http://dygraphs.com/
func getGraphHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html")
	vars := mux.Vars(r)
	status, latency, err := datastore.Read(vars["ief"])
	if err != nil {
		log.Println("Failed to read from DB:", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = fmt.Fprintf(w, plotsTemplateHead, asset(dygraphs))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf := make([]byte, 0, 128)
	for t, value := range latency {
		buf = append(buf, t.Format(time.RFC3339Nano)...)
		buf = append(buf, ","...)

		if status[t] {
			buf = append(buf, "NaN,"...)
			buf = append(buf, strconv.FormatFloat(value, 'f', -1, 32)...)
		} else {
			buf = append(buf, strconv.FormatFloat(value, 'f', -1, 32)...)
			buf = append(buf, ",NaN"...)
		}
		buf = append(buf, `\n`...)

		_, err = w.Write(buf)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		buf = buf[:0]
	}

	_, err = fmt.Fprintf(w, plotsTemplateTail, vars["ief"])
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func InitServer() error {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/v1/{ief}", postDataHandler).Methods("POST")
	router.HandleFunc("/v1/{ief}", getGraphHandler).Methods("GET")
	bind := fmt.Sprintf("%s:%d", config.C.ListenHost, config.C.ListenPort)
	log.Println("listening on: ", bind)
	return http.ListenAndServe(bind, router)
}
