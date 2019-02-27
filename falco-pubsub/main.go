package main

import (
	"bufio"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"log"
	"os"
	"regexp"
	"sync"
	"time"
)

var (
	googleProjectID       = os.Getenv("GOOGLE_PROJECT_ID")
	googleCredentialsData = os.Getenv("GOOGLE_CREDENTIALS_DATA")
	googleTopicName       = os.Getenv("GOOGLE_TOPIC_NAME")
	slugRegularExpression = regexp.MustCompile("[^a-z0-9]+")
)

func main() {
	var pipePath = flag.String("f", "/var/run/falco/nats", "The named pipe path")

	if googleProjectID == "" || googleCredentialsData == "" {
		log.Fatalln("You need to provide the env vars GOOGLE_PROJECT_ID and GOOGLE_CREDENTIALS_DATA")
	}

	credentials, err := google.CredentialsFromJSON(context.Background(), []byte(googleCredentialsData), storage.ScopeReadOnly)
	if err != nil {
		log.Fatalf("could not create credentials from json data: %v", err)
	}

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	pubsubClient, err := pubsub.NewClient(context.Background(), googleProjectID, option.WithCredentials(credentials))
	if err != nil {
		log.Fatalf("could not create a new client for Google Project ID '%s': %v", googleProjectID, err)
	}
	defer pubsubClient.Close()

	log.Printf("Created new client for Google Project ID: %s", googleProjectID)

	_, _ = pubsubClient.CreateTopic(context.Background(), googleTopicName)

	topic := pubsubClient.Topic(googleTopicName)
	defer topic.Stop()
	log.Printf("Using topic '%s'", googleTopicName)

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
