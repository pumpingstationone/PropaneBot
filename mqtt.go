package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type CurrentData struct {
	Weight    string
	TimeStamp string
	// We'll calculate this when setting so the
	// bot doesn't have to
	Remaining string
}

// We're not being fancy about this, we only care about the
// current value, so we're just going to store it here so that
// it's available when someone asks via the bot. We *do* guard
// the writes though, we're not complete animals. :)
var currentData CurrentData

var mutex = &sync.Mutex{}

func niceDate(unixts string) string {
	i, err := strconv.ParseInt(unixts, 10, 64)
	if err != nil {
		panic(err)
	}
	tm := time.Unix(i, 0)

	local := tm
	// PS1 is located in Chicago
	location, err := time.LoadLocation("America/Chicago")
	if err == nil {
		local = local.In(location)
	}

	return local.Format("Mon Jan _2 03:04PM 2006")
}

func onMessageReceived(client MQTT.Client, message MQTT.Message) {
	//fmt.Printf("Received message on topic: %s\nMessage: %s\n", message.Topic(), message.Payload())

	// payload is in the format: 1577640142,163.4
	payload := message.Payload()
	parts := strings.Split(string(payload), ",")

	datePart := niceDate(parts[0])
	weightPart := parts[1]
	percentRemaining := calcRemaining(weightPart)

	mutex.Lock()
	currentData.Weight = weightPart
	currentData.TimeStamp = datePart
	currentData.Remaining = percentRemaining
	mutex.Unlock()
}

func listenOnTopic() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	hostname, _ := os.Hostname()

	server := "" // The mqtt server name (with port)
	topic := ""  // Whatever the topic is
	qos := 0
	clientid := hostname

	connOpts := MQTT.NewClientOptions().AddBroker(server).SetClientID(clientid).SetCleanSession(true)

	connOpts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(topic, byte(qos), onMessageReceived); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}

	client := MQTT.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		log.Printf("Connected to %s\n", server)
	}

	<-c
}
