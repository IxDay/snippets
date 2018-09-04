package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	std = log.New(os.Stderr, "", log.LstdFlags)
)

func handler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		std.Println(err)
	}
	w.Write(body)
	std.Printf("%s %s %s", r.Method, r.URL, body)
}

func main() {
	var host string
	var port int

	flag.StringVar(&host, "h", "localhost", "Hostname from which the server will serve request")
	flag.IntVar(&port, "p", 5678, "Port on which the server will serve request")
	flag.Parse()

	server := &http.Server{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Handler:  http.HandlerFunc(handler),
		ErrorLog: std,
	}
	std.Printf("Serving on %s:%d...", host, port)
	std.Fatalln(server.ListenAndServe())
}
