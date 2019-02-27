package main

import (
	"bufio"
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
	"time"
)

const (
	TOPIC_NAME = "topic-name"
)

var (
	googleProjectID       = os.Getenv("GOOGLE_PROJECT_ID")
	slugRegularExpression = regexp.MustCompile("[^a-z0-9]+")
)

func main() {
	var pipePath = flag.String("f", "/var/run/falco/nats", "The named pipe path")

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	pubsubClient, err := pubsub.NewClient(context.Background(), googleProjectID)
	if err != nil {
		log.Fatalf("could not create a new client for Google Project ID '%s': %v", googleProjectID, err)
	}
	defer pubsubClient.Close()

	log.Printf("Created new client for Google Project ID: %s", googleProjectID)

	_, _ = pubsubClient.CreateTopic(context.Background(), TOPIC_NAME)

	topic := pubsubClient.Topic(TOPIC_NAME)
	defer topic.Stop()
	log.Printf("Using topic '%s'", TOPIC_NAME)

	pipe, err := os.OpenFile(*pipePath, os.O_RDONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Opened pipe %s", *pipePath)

	reader := bufio.NewReader(pipe)
	scanner := bufio.NewScanner(reader)
	log.Printf("Scanning %s", *pipePath)

	wg := &sync.WaitGroup{}

	for scanner.Scan() {
		msg := []byte(scanner.Text())
		wg.Add(1)
		go publish(topic, msg, wg)
	}

	wg.Wait()

}

func publish(topic *pubsub.Topic, msg []byte, wg *sync.WaitGroup) {
	defer wg.Done()

	alert := parseAlert(msg)

	message := &pubsub.Message{
		Data: msg,
		Attributes: map[string]string{
			"priority": alert.Priority,
			"rule":     alert.Rule,
		},
	}

	ctxPublish, cancelPublish := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelPublish()

	result := topic.Publish(ctxPublish, message)

	ctxGet, cancelGet := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelGet()

	id, err := result.Get(ctxGet)
	if err != nil {
		log.Fatalf("error publishing message: %v")
	}
	fmt.Printf("Published a message with a message ID: %s\n", id)
}

func usage() {
	log.Fatalf("Usage: %s [-s server] <subject> <msg> \n", os.Args[0])
}

type parsedAlert struct {
	Priority string `json:"priority"`
	Rule     string `json:"rule"`
}

func parseAlert(alert []byte) *parsedAlert {
	var result parsedAlert
	err := json.Unmarshal(alert, &result)
	if err != nil {
		log.Fatal(err)
	}

	return &result
}
