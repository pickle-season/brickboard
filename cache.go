package main

import (
	"fmt"
	"os"
	"time"
)

func (c *DataCache) Start(interval time.Duration) {
	fmt.Println("Starting cache update routine...")
	go func() {
		fmt.Println("Getting DSM auth...")
		dsm_auth := dsmLogin(os.Getenv("DSM_ACCOUNT"), os.Getenv("DSM_PASSWORD"))
		fmt.Printf("DSM Auth: %+v\n", dsm_auth)

		for {
			fmt.Println("Updating data cache...")

			ha_states, err := getHaStateMap()
			if err != nil {
				panic(err)
			}

			dsm_storage := dsmGetStorage(dsm_auth)

			pc_states := getPcStates()

			c.mu.Lock()
			c.data = &TemplateData{HaStateMap: ha_states, DsmStorage: dsm_storage, PcOnline: pc_states}
			c.mu.Unlock()

			time.Sleep(interval)
		}
	}()
}

func (c *DataCache) Get() *TemplateData {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data
}
