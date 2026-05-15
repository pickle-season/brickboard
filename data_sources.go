package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

const HA_API_URL = "https://ha.krabice.online/api/"
const DSM_API_URL = "https://kostka-cukru.lan:5001/webapi/entry.cgi"
const PVE_API_URL = "https://pve.lan:8006/api2/json/"

func ping(host string) bool {
	fmt.Printf("Starting ping for %s\n", host)
	pinger, err := probing.NewPinger(host)
	if err != nil {
		fmt.Printf("Ping failed: %+v\n", err)
		return false
	}
	pinger.Count = 3
	pinger.Timeout = time.Second * 10
	if os.Getenv("SOCK_PRIV") == "1" {
		pinger.SetPrivileged(true)
	}
	err = pinger.Run() // Blocks until finished.
	if err != nil {
		fmt.Printf("Ping failed: %+v\n", err)
		return false
	}
	stats := pinger.Statistics() // get send/receive/duplicate/rtt stats
	fmt.Printf("Ping stats: %+v\n", stats)
	return stats.PacketsRecv > 0
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

func dsmGetStorage(auth DsmAuth) DsmStorage {
	fmt.Println("Getting DSM storage info...")

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

	dsmStorage.Data.Volumes[0].Size.Used = fmt.Sprintf("%v", usedGB)
	dsmStorage.Data.Volumes[0].Size.Total = fmt.Sprintf("%v", totalGB)

	fmt.Println("DSM Storage:", dsmStorage)

	return dsmStorage
}

func getHaStateMap() (HaStateMap, error) {
	fmt.Println("Updating HA state map...")

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

	fmt.Printf("temp: %+v\nhumidity: %+v\nco2: %+v\nexternal_ip: %+v\n", temperature, humidity, co2, external_ip)
	return HaStateMap{
		Temperature: temperature,
		Humidity:    humidity,
		Co2:         co2,
		ExternalIp:  external_ip,
	}, nil
}

func getPcStates() []PcState {
	return []PcState{
		{Name: "cat-heater", Online: ping("cat-heater.lan")},
		{Name: "bad-boi", Online: ping("bad-boi.lan")},
		{Name: "media-box", Online: ping("media-box.lan")},
	}
}

func getPveMetrics() PveMetrics {
	fmt.Println("Updating PVE Metrics...")

	req, err := http.NewRequest("GET", fmt.Sprintf("%scluster/resources?type=node", PVE_API_URL), nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", os.Getenv("PVE_AUTH"))
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var pveMetrics PveMetrics

	json.NewDecoder(resp.Body).Decode(&pveMetrics)

	for i, node := range pveMetrics.Data {
		// pveMetrics.Data[i].Cpu = node.Cpu * 100
		// pveMetrics.Data[i].Maxcpu = node.Maxcpu * 100
		// pveMetrics.Data[i].Mem = node.Mem / (1024 * 1024 * 1024)
		// pveMetrics.Data[i].Maxmem = node.Maxmem / (1024 * 1024 * 1024)
		// pveMetrics.Data[i].Disk = node.Disk / (1024 * 1024 * 1024)
		// pveMetrics.Data[i].Maxdisk = node.Maxdisk / (1024 * 1024 * 1024)
		fmt.Printf("Node: %s, CPU: %.2f%%, Mem: %dGB, Disk: %dGB\n", node.Node, node.Cpu, pveMetrics.Data[i].Mem, pveMetrics.Data[i].Disk)
	}

	return pveMetrics
}
