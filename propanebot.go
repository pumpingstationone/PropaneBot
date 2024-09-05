package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

// func getenv(name string) string {
// 	v := os.Getenv(name)
// 	if v == "" {
// 		panic("missing required environment variable " + name)
// 	}
// 	return v
// }

type AppConfig struct {
	MQTT struct {
		Server string `json:"server"`
		Topic  string `json:"topic"`
	} `json:"mqtt"`
	Discord struct {
		AppToken string `json:"appToken"`
		GuildID  string `json:"guildId"`
		BotToken string `json:"botToken"`
	} `json:"discord"`
	Slack struct {
		APIToken string `json:"apiToken"`
	} `json:"slack"`
}

func main() {
	// Let's begin by reading the cylinder settings
	LoadCylinderData()

	ds := NewDatastore()
	var cfg AppConfig
	if err := LoadConfig("./config.json", &cfg); err != nil {
		panic("Failed to load config: " + err.Error())
	}
	// Get a Context that can handle stopping for signals, timeouts, or whatever else we throw at it
	ctx, done := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer done()
	wg, ctx := errgroup.WithContext(ctx)

	// Now start the mqtt stuff so we can start getting messages
	wg.Go((&MQTTListener{
		Datastore: ds,
		Server:    cfg.MQTT.Server,
		Topic:     cfg.MQTT.Topic,
	}).Run(ctx))

	// Setup and run Discord
	wg.Go((&DiscordBot{
		AppToken:  cfg.Discord.AppToken,
		GuildID:   cfg.Discord.GuildID,
		BotToken:  cfg.Discord.BotToken,
		Datastore: ds,
	}).Run(ctx))

	// Now begins the Slack stuff
	wg.Go((&SlackBot{
		APIToken:  cfg.Slack.APIToken,
		Datastore: ds,
	}).Run(ctx))

	// Wait for exit and print any error messages that bubble up
	log.Printf("Exiting with message: %q\n", wg.Wait())
}

func LoadConfig(path string, cfg *AppConfig) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(cfg)
}
