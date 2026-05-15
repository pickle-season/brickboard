package main

//go:generate go run saml.dev/gome-assistant/cmd/generate
import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"crypto/tls"
	"encoding/json"

	"strconv"

	"github.com/joho/godotenv"
	probing "github.com/prometheus-community/pro-bing"
)

const HA_API_URL = "https://ha.krabice.online/api/"
const DSM_API_URL = "https://kostka-cukru:5001/webapi/entry.cgi"

type DataCache struct {
	mu   sync.RWMutex
	data *TemplateData
}

type HaStateMap struct {
	Temperature HAState
	Humidity    HAState
	Co2         HAState
	ExternalIp  HAState
}

type HAState struct {
	EntityId   string `json:"entity_id"`
	State      string `json:"state"`
	Attributes struct {
		UnitOfMeasurement string `json:"unit_of_measurement"`
	} `json:"attributes"`
}

type DsmAuth struct {
	Data struct {
		Did string `json:"did"`
		Sid string `json:"sid"`
	} `json:"data"`
	Success bool `json:"success"`
}

type PcOnline struct {
	Cat_heater bool
	Bad_boi    bool
}

type TemplateData struct {
	HaStateMap HaStateMap
	DsmStorage DsmStorage
	PcOnline   PcOnline
}

var tr = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

var client = &http.Client{Transport: tr}

func makeHttp(ha_state_map HaStateMap, dsm_storage DsmStorage) string {
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

func (c *DataCache) Get() *TemplateData {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data
}

func ping(host string) bool {
	pinger, err := probing.NewPinger(host)
	if err != nil {
		return false
	}
	pinger.Count = 3
	pinger.Timeout = time.Second * 5
	err = pinger.Run() // Blocks until finished.
	if err != nil {
		return false
	}
	stats := pinger.Statistics() // get send/receive/duplicate/rtt stats
	fmt.Printf("Ping stats: %+v\n", stats)
	return stats.PacketsRecv > 0
}

func getPcOnline() PcOnline {
	return PcOnline{
		Cat_heater: ping("cat-heater"),
		Bad_boi:    ping("bad-boi"),
	}
}

func (c *DataCache) Start(interval time.Duration) {
	go func() {
		dsm_auth := dsmLogin(os.Getenv("DSM_ACCOUNT"), os.Getenv("DSM_PASSWORD"))
		fmt.Printf("DSM Auth: %+v\n", dsm_auth)

		for {
			fmt.Println("Updating data cache...")

			ha_states, err := getHaStateMap()
			if err != nil {
				panic(err)
			}

			dsm_storage := dsmGetStorage(dsm_auth)

			pc_online := getPcOnline()

			c.mu.Lock()
			c.data = &TemplateData{HaStateMap: ha_states, DsmStorage: dsm_storage, PcOnline: pc_online}
			c.mu.Unlock()

			time.Sleep(interval)
		}
	}()
}

func getHaState(entityId string) (HAState, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%sstates/%s", HA_API_URL, entityId), nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("HA_AUTH_TOKEN")))
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var haState HAState

	json.NewDecoder(resp.Body).Decode(&haState)
	return haState, nil
}

func dsmLogin(account string, passwd string) DsmAuth {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s?api=SYNO.API.Auth&version=6&method=login&account=%s&passwd=%s", DSM_API_URL, account, passwd), nil)
	if err != nil {
		panic(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var dsm_resp DsmAuth

	json.NewDecoder(resp.Body).Decode(&dsm_resp)
	return dsm_resp
}

type DsmStorage struct {
	Data struct {
		Volumes []struct {
			Size struct {
				Total string `json:"total"`
				Used  string `json:"used"`
			} `json:"size"`
		} `json:"volumes"`
	} `json:"data"`
	Success bool `json:"success"`
}

func dsmGetStorage(auth DsmAuth) DsmStorage {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s?api=SYNO.Storage.CGI.Storage&version=1&method=load_info&_sid=%s", DSM_API_URL, auth.Data.Sid), nil)
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	var dsmStorage DsmStorage
	json.NewDecoder(resp.Body).Decode(&dsmStorage)

	used, err := strconv.Atoi(dsmStorage.Data.Volumes[0].Size.Used)
	if err != nil {
		panic(err)
	}
	total, err := strconv.Atoi(dsmStorage.Data.Volumes[0].Size.Total)
	if err != nil {
		panic(err)
	}

	usedGB := used / (1024 * 1024 * 1024)
	totalGB := total / (1024 * 1024 * 1024)

	dsmStorage.Data.Volumes[0].Size.Used = fmt.Sprintf("%v GB", usedGB)
	dsmStorage.Data.Volumes[0].Size.Total = fmt.Sprintf("%v GB", totalGB)

	return dsmStorage
}

func getHaStateMap() (HaStateMap, error) {
	temperature, err := getHaState("sensor.jesus_sensor_living_room_temperature")
	if err != nil {
		return HaStateMap{}, err
	}
	humidity, err := getHaState("sensor.jesus_sensor_living_room_humidity")
	if err != nil {
		return HaStateMap{}, err
	}
	co2, err := getHaState("sensor.jesus_sensor_living_room_co2")
	if err != nil {
		return HaStateMap{}, err
	}
	external_ip, err := getHaState("sensor.archer_ax23_external_ip")
	if err != nil {
		return HaStateMap{}, err
	}

	return HaStateMap{
		Temperature: temperature,
		Humidity:    humidity,
		Co2:         co2,
		ExternalIp:  external_ip,
	}, nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	c := &DataCache{}
	c.Start(time.Duration(5) * time.Second)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("Content-Type", "text/css")
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
		//w.Write([]byte(makeHttp(ha_states, dsm_storage)))

	})

	//entities.Sensor.JesusSensorLivingRoomTemperature
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		//http.ServeFile(w, r, filepath.Join("res", "index.html"))

		// ha_states, err := getHaStateMap()
		// if err != nil {
		// 	panic(err)
		// }

		// dsm_storage := dsmGetStorage(dsm_auth)

		//w.Write([]byte(makeHttp(ha_states, dsm_storage)))
		http.ServeFile(w, r, filepath.Join("static", "index.html"))
	})

	http.ListenAndServe("0.0.0.0:3000", r)
}
