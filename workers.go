package main

import (
	"fmt"
	"log"
	"time"
)

func startPurgeRefresh() {
	go func() { // Daily purge of expired refresh tokens
		for {
			err := deleteExpiredRefresh()
			if err != nil {
				log.Println(fmt.Sprintf("Failed to purge refresh: %s", err.Error()))
			}
			time.Sleep(24 * time.Hour)
		}
	}()
}
