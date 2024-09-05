package main

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	// From the Discord Developer Portal
	AppToken string
	// Optional. Restricts bot to 1 server. Aka Server ID. In Discord, enable "Developer Mode", then right-click on the the server's icon
	GuildID string
	// From the Discord Developer Portal within an app
	BotToken  string
	Datastore *Datastore
	session   *discordgo.Session
}

func (b *DiscordBot) Run(ctx context.Context) func() error {
	return func() error {
		var err error
		b.session, err = discordgo.New("Bot " + b.BotToken)
		if err != nil {
			return err
		}
		b.session.AddHandler(b.handleReady())
		b.session.AddHandler(b.handleWeight())
		if _, err := b.session.ApplicationCommandBulkOverwrite(b.AppToken, b.GuildID, b.buildCommands()); err != nil {
			return err
		}
		if err := b.session.Open(); err != nil {
			return err
		}
		defer func() { err = b.session.Close() }()
		<-ctx.Done()
		log.Printf("DiscordBot received Done with Error %q. Shutting down.\n", ctx.Err())
		return err
	}
}

func (b *DiscordBot) buildCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "weight",
			Description: "Get the current propane level",
			// Options: []*discordgo.ApplicationCommandOption{},
		},
	}
}

func (b *DiscordBot) handleReady() func(*discordgo.Session, *discordgo.Ready) {
	return func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Printf("Bot started as: %q", r.User.String())
	}
}

func (b *DiscordBot) handleWeight() func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}
		data := i.ApplicationCommandData()
		if data.Name != "weight" {
			return
		}
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: b.Datastore.GetString()},
		}); err != nil {
			fmt.Printf("Error: Failed to send response: %s", err)
		}
	}
}
