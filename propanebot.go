package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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
	mu sync.Mutex `json:"-"`

	MQTT struct {
		Server string `json:"server"`
		Topic  string `json:"topic"`
	} `json:"mqtt"`
	Discord struct {
		AppToken  string `json:"appToken"`
		GuildID   string `json:"guildId"`
		BotToken  string `json:"botToken"`
		ChannelID string `json:"channelId"`
		UserID    string `json:"userId"`
	} `json:"discord"`
	Slack struct {
		APIToken string `json:"apiToken"`
	} `json:"slack"`
	NotificationSent bool `json:"notificationSent"`
}

func (cfg *AppConfig) HasSentNotification() bool {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	return cfg.NotificationSent
}

func (cfg *AppConfig) SetNotificationSent(path string, sent bool) error {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	cfg.NotificationSent = sent
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func LoadConfig(path string, cfg *AppConfig) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(cfg)
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

	// Watch cylinder.json so edits (including from the web settings page)
	// are picked up without restarting the bot
	wg.Go(WatchCylinderData(ctx))

	// Now start the mqtt stuff so we can start getting messages
	wg.Go((&MQTTListener{
		Datastore: ds,
		Server:    cfg.MQTT.Server,
		Topic:     cfg.MQTT.Topic,
	}).Run(ctx))

	// Setup and run Discord
	dc := &DiscordBot{AppToken: cfg.Discord.AppToken,
		GuildID:   cfg.Discord.GuildID,
		BotToken:  cfg.Discord.BotToken,
		ChannelID: cfg.Discord.ChannelID,
		UserID:    cfg.Discord.UserID,
		Datastore: ds}
	wg.Go(dc.Run(ctx))

	// Setup and run the propane monitor that will send alerts to Discord when the level is low
	monitor := NewPropaneMonitor(dc, ds, 10*time.Second)
	monitor.Config = &cfg
	monitor.ConfigPath = "./config.json"
	go monitor.Start(ctx)

	// Start the web server on port 9991
	wg.Go((&WebServer{
		Port:       9991,
		Datastore:  ds,
		Config:     &cfg,
		ConfigPath: "./config.json",
	}).Run(ctx))

	// Wait for exit and print any error messages that bubble up
	log.Printf("Exiting with message: %q\n", wg.Wait())
}
