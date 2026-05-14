package main

//go:generate go run saml.dev/gome-assistant/cmd/generate
import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/joho/godotenv"
)

const HA_API_URL = "https://ha.krabice.online/api/"

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s", HA_API_URL), nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("HA_AUTH_TOKEN")))
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	//entities.Sensor.JesusSensorLivingRoomTemperature
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		//http.ServeFile(w, r, filepath.Join("res", "index.html"))

		w.Write([]byte("brickboard: welcome"))
	})
	http.ListenAndServe("192.168.0.11:3000", r)
}
