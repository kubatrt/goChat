package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"

	"github.com/matryer/goblueprints/chapter1/trace"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, r)
}

const version = 0.2

func main() {
	var addr = flag.String("addr", "localhost:8080", "http service address")
	flag.Parse()

	//u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}

	r := newRoom()
	r.tracer = trace.New(os.Stdout)

	t := &templateHandler{filename: "chat.html"}
	http.Handle("/", t)
	http.Handle("/room", r)

	go r.run()

	log.Println("Starting chat server:", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
