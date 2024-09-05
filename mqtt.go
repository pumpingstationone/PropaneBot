package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type MQTTListener struct {
	Datastore *Datastore
	// The MQTT server's name with port
	Server    string
	// The topic to listen on
	Topic     string
	// Cylinder  Cylinder
}

func (l *MQTTListener) niceDate(unixts string) time.Time {
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
	return local
}

func (l *MQTTListener) parseWeight(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	return f
}

func (l *MQTTListener) onMessageReceived(client MQTT.Client, message MQTT.Message) {
	//fmt.Printf("Received message on topic: %s\nMessage: %s\n", message.Topic(), message.Payload())

	// payload is in the format: 1577640142,163.4
	payload := message.Payload()
	parts := strings.Split(string(payload), ",")
	weight := l.parseWeight(parts[1])

	l.Datastore.Set(
		weight,
		l.niceDate(parts[0]),
		cylinder.CalcRemaining(weight),
	)
}

// Initialize and start the MQTTListener
func (l *MQTTListener) Run(ctx context.Context) func() error {
	return func() error {
		hostname, _ := os.Hostname()
		qos := 0
		clientid := hostname

		connOpts := MQTT.NewClientOptions().AddBroker(l.Server).SetClientID(clientid).SetCleanSession(true)

		connOpts.OnConnect = func(c MQTT.Client) {
			if token := c.Subscribe(l.Topic, byte(qos), l.onMessageReceived); token.Wait() && token.Error() != nil {
				panic(token.Error())
			}
		}

		client := MQTT.NewClient(connOpts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			panic(token.Error())
		} else {
			log.Printf("Connected to %s\n", l.Server)
		}

		<-ctx.Done()
		log.Printf("MQTT received Done with Error %q. Shutting down.\n", ctx.Err().Error())

		return nil
	}
}