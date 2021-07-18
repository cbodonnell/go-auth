package main

import (
	"fmt"
	"log"
	"time"
)

// Daily purge of expired refresh tokens
func startPurgeRefresh() {
	go func() {
		for {
			err := deleteExpiredRefresh()
			if err != nil {
				log.Println(fmt.Sprintf("Failed to purge refresh: %s", err.Error()))
			}
			time.Sleep(24 * time.Hour)
		}
	}()
}
