package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

var config Config

func main() {
	log.Println("started")
	log.SetPrefix("nilsu docs: ")
	if len(os.Args) > 1 {
		if err := config.Load(os.Args[1]); err != nil {
			log.Fatalln("failed to read config")
		}
		log.Println("config loaded")
	} else {
		if err := config.Load("/etc/nilsu/docs.toml"); err != nil {
			log.Fatalln("failed to read config")
		}
		log.Println("config loaded")
	}

	// create router
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(config.StaticDir))))
	r.HandleFunc("/", IndexHandler)
	r.HandleFunc("/create", CreateHandler)
	r.HandleFunc("/save", SaveHandler)
	r.HandleFunc("/view/{name}", ViewLatestHandler)
	r.HandleFunc("/view/{name}/{time:[0-9]+}", ViewTimeHandler)
	r.HandleFunc("/read/{hash}", ReadLatestHandler)
	r.HandleFunc("/read/{hash}/{time:[0-9]+}", ReadTimeHandler)
	r.HandleFunc("/edit/{name}", EditHandler)

	// starting up the server
	log.Printf("serving %s\n", config.Address)
	srv := &http.Server{
		Addr:           config.Address,
		Handler:        r,
		ReadTimeout:    time.Duration(config.ReadTimeout * int64(time.Second)),
		WriteTimeout:   time.Duration(config.WriteTimeout * int64(time.Second)),
		MaxHeaderBytes: 1 << 20,
	}
	srv.ListenAndServe()
}
