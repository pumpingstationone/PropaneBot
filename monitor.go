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
	Config         *AppConfig
	ConfigPath     string
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
				if !pm.Config.HasSentNotification() {
					message := fmt.Sprintf("Hey <@%s>! The cylinder has dropped below %.0f%%! Current level: %.2f%%.\nMight wanna think about ordering a new one.", pm.discordClient.UserID, pm.alertThreshold, currentLevel)

					// Send notification to your specific Discord channel/user
					err := pm.discordClient.SendMessage(message)
					if err != nil {
						log.Printf("Failed to send Discord alert: %v\n", err)
					} else {
						log.Println("Discord alert sent successfully.")
						if err := pm.Config.SetNotificationSent(pm.ConfigPath, true); err != nil {
							log.Printf("Failed to persist Discord notification state: %v\n", err)
						}
					}
				}
			}
		}
	}
}
