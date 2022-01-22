package main

import (
	"fmt"
	"log"
	"time"
)

// Daily purge of expired refresh tokens
func startPurgeRefresh() {
	for {
		log.Println("purging expired refresh tokens")
		err := deleteExpiredRefresh()
		if err != nil {
			log.Println(fmt.Sprintf("failed to purge refresh: %s", err.Error()))
		}
		time.Sleep(24 * time.Hour)
	}
}
