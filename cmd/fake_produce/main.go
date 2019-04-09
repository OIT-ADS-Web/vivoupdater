package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/OIT-ADS-Web/vivoupdater"
	"github.com/OIT-ADS-Web/vivoupdater/kafka"
	"github.com/Shopify/sarama"
)

var servers []string
var config *sarama.Config

func init() {
	servers = strings.Split(os.Getenv("BOOTSTRAP_SERVERS"), ",")

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("pwd=%v\n", dir)

	tlsConfig, err := kafka.NewTLSConfig(
		fmt.Sprintf("%s", os.Getenv("CLIENT_CERT")),
		fmt.Sprintf("%s", os.Getenv("CLIENT_KEY")),
		fmt.Sprintf("%s", os.Getenv("SERVER_CERT")))

	if err != nil {
		log.Fatalf("could not read cert file: %v\n", err)
	}
	config = sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	config.Consumer.Return.Errors = true
	//config.Consumer.MaxWaitTime = time.Duration(305000 * time.Millisecond)
	config.ClientID = "rn47"

	// offset makes a difference
	//config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Net.TLS.Config = tlsConfig
	config.Net.TLS.Enable = true
	config.Version = sarama.V1_0_0_0

}

func main() {
	flag.Parse()

	producer, err := sarama.NewSyncProducer(servers, config)
	if err != nil {
		log.Fatalf("%s\n", err)
		log.Fatalln(err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalf("%s\n", err)
			log.Fatalln(err)
		}
	}()

	topic := os.Getenv("TOPICS")

	num := rand.Intn(9)
	subject := fmt.Sprintf("http://scholars.duke.edu/individual/per000000%d", num)

	um := vivoupdater.UpdateMessage{
		Type:  "update",
		Phase: "destination",
		Name:  "urn:x-arq:UnionGraph",
		Triple: vivoupdater.Triple{
			Subject:   subject,
			Predicate: "rdfs:label",
			Object:    fmt.Sprintf("Lester the Tester #%d", num)},
	}

	val, err := json.Marshal(um)
	if err != nil {
		log.Fatalf("%s\n", err)
		log.Printf("FAILED to json marshal: %s\n", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(string(val))}
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		log.Fatalf("%s\n", err)
		log.Printf("FAILED to send message: %s\n", err)
	} else {
		log.Printf("subject=%s\n", subject)
		log.Printf("> message sent to partition %d at offset %d\n", partition, offset)
	}

}
