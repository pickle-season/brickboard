package main

//go:generate go run saml.dev/gome-assistant/cmd/generate
import (
	"bytes"
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"crypto/tls"

	"github.com/joho/godotenv"
)

const HA_API_URL = "https://ha.krabice.online/api/"
const DSM_API_URL = "https://kostka-cukru.lan:5001/webapi/entry.cgi"

var tr = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

var client = &http.Client{Transport: tr}

func _makeHttp(ha_state_map HaStateMap, dsm_storage DsmStorage) string {
	tmpl, err := template.ParseFiles(filepath.Join("static", "templates", "index.gohtml"))
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, TemplateData{HaStateMap: ha_state_map, DsmStorage: dsm_storage})
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	c := &DataCache{}
	c.Start(time.Duration(10) * time.Second)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
	})
	r.Get("/static/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, filepath.Join("static", "style.css"))
	})
	r.Get("/static/main.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		http.ServeFile(w, r, filepath.Join("static", "main.js"))
	})

	r.Get("/data", func(w http.ResponseWriter, r *http.Request) {
		data := c.Get()
		render.JSON(w, r, data)
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, filepath.Join("static", "index.html"))
	})

	http.ListenAndServe("0.0.0.0:3001", r)
}
