package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/slack-go/slack"
)

var slackWeightRegexp = regexp.MustCompile("!weight")

type SlackBot struct {
	APIToken  string
	Datastore *Datastore
	rtm       *slack.RTM
}

func (b *SlackBot) Run(ctx context.Context) func() error {
	return func() error {
	Loop:
		for {
			select {
			case msg := <-b.getRTM().IncomingEvents:
				switch ev := msg.Data.(type) {
				case *slack.MessageEvent:
					//info := rtm.GetInfo()

					text := ev.Text
					text = strings.TrimSpace(text)
					text = strings.ToLower(text)

					// Let's see if someone asked us for something...
					sendResponse, response := b.checkForCommands(text)

					if sendResponse {
						// ...yep, we sent something back, so let's send it to the channel
						if err := b.SendMessage(ctx, ev.Channel, []byte(response)); err != nil {
							log.Printf("Failed to respond to weight request with: %s\n", err)
						}
					}

				case *slack.RTMError:
					fmt.Printf("Error: %s\n", ev.Error())

				case *slack.InvalidAuthEvent:
					fmt.Printf("Invalid credentials")
					break Loop

				default:
					// Nothin' to do
				}
			case <-ctx.Done():
				log.Printf("SlackBot received Done with Error %q. Shutting down.\n", ctx.Err())
				return nil
			}
		}
		return nil
	}
}

func (b *SlackBot) SendMessage(ctx context.Context, channel string, msg []byte) error {
	_, _, _, err := b.getRTM().SendMessageContext(ctx, channel, slack.MsgOptionText(string(msg), false))
	return err
}

func (b *SlackBot) getRTM() *slack.RTM {
	if b.rtm == nil {
		api := slack.New(b.APIToken)
		b.rtm = api.NewRTM()
		go b.rtm.ManageConnection()
	}
	return b.rtm
}

func (b *SlackBot) checkForCommands(input string) (bool, string) {
	if slackWeightRegexp.MatchString(input) {
		return true, b.Datastore.GetString()
	}
	return false, ""
}
