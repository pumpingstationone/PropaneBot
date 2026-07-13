package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

// PropaneMonitor manages the background check loop
type PropaneMonitor struct {
	discordClient  *DiscordBot // Your Discord bot/webhook client
	datastore      *Datastore  // Component that reads the cylinder/propane value
	checkInterval  time.Duration
	alertThreshold float64
}

func NewPropaneMonitor(dc *DiscordBot, ds *Datastore, interval time.Duration) *PropaneMonitor {
	return &PropaneMonitor{
		discordClient:  dc,
		datastore:      ds,
		checkInterval:  interval,
		alertThreshold: 20.0,
	}
}

// Start runs the monitoring loop in a background thread
func (pm *PropaneMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(pm.checkInterval)
	defer ticker.Stop()

	log.Println("Background propane monitor started...")
	alertSent := false

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping propane monitor...")
			return
		case <-ticker.C:
			// Fetch the current percentage level from the datastore
			currentLevel := pm.datastore.Get().Remaining

			log.Printf("Current propane level: %.2f%%\n", currentLevel)

			// Alert condition
			if currentLevel < pm.alertThreshold {
				if !alertSent {
					message := fmt.Sprintf("Hey @tachoknight! The cylinder has dropped below %.0f%%! Current level: %.2f%%.\nMight wanna think about ordering a new one.", pm.alertThreshold, currentLevel)

					// Send notification to your specific Discord channel/user
					err := pm.discordClient.SendMessage(message)
					if err != nil {
						log.Printf("Failed to send Discord alert: %v\n", err)
					} else {
						log.Println("Discord alert sent successfully.")
						alertSent = true
					}
				}
			} else {
				// Reset the alert state once the tank is refilled above the threshold
				if alertSent {
					log.Println("Propane levels restored above threshold. Resetting alert trigger.")
					alertSent = false
				}
			}
		}
	}
}
