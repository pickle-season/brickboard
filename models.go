package main

import "sync"

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

type PcState struct {
	Name   string
	Online bool
}

type TemplateData struct {
	HaStateMap HaStateMap
	DsmStorage DsmStorage
	PcOnline   []PcState
	PveMetrics PveMetrics
}

type PveMetrics struct {
	Data []struct {
		Node    string  `json:"node"`
		Cpu     float64 `json:"cpu"`
		Maxcpu  float64 `json:"maxcpu"`
		Mem     int     `json:"mem"`
		Maxmem  int     `json:"maxmem"`
		Disk    int     `json:"disk"`
		Maxdisk int     `json:"maxdisk"`
	} `json:"data"`
}
