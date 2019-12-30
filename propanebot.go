package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/nlopes/slack"
)

func getenv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		panic("missing required environment variable " + name)
	}
	return v
}

func checkForCommands(input string) (bool, string) {
	response := ""
	sendResponse := false

	matched, _ := regexp.MatchString("!weight", input)
	if matched {
		sendResponse = true
		mutex.Lock()
		response = fmt.Sprintf("Well, as of %s the cylinder weighs %s lbs which kinda translates into %s remaining", currentData.TimeStamp, currentData.Weight, currentData.Remaining)
		mutex.Unlock()
	}

	return sendResponse, response
}

func main() {
	// Let's begin by reading the cylinder settings
	loadCylinderData()

	// Now start the mqtt stuff so we can start getting messages
	go listenOnTopic()

	//
	// Now begins the Slack stuff
	//
	token := "" // ToDo: Need the appropriate API token from Slack
	api := slack.New(token)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				//info := rtm.GetInfo()

				text := ev.Text
				text = strings.TrimSpace(text)
				text = strings.ToLower(text)

				// Let's see if someone asked us for something...
				sendResponse, response := checkForCommands(text)

				if sendResponse {
					// ...yep, we sent something back, so let's send it to the channel
					rtm.SendMessage(rtm.NewOutgoingMessage(response, ev.Channel))
				}

			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				fmt.Printf("Invalid credentials")
				break Loop

			default:
				// Nothin' to do
			}
		}
	}
}
