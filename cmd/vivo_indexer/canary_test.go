package main_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/OIT-ADS-Web/vivoupdater/kafka"
	"github.com/Shopify/sarama"
)

var servers []string
var config *sarama.Config

// FIXME: don't know how to import these (circular)
type Triple struct {
	Subject   string
	Predicate string
	Object    string
}

type UpdateMessage struct {
	Type   string
	Phase  string
	Name   string
	Triple Triple
}

func setup() {
	servers = strings.Split(os.Getenv("BOOTSTRAP_SERVERS"), ",")
	//caCert, err := ioutil.ReadFile(os.Getenv("CLIENT_CERT"))

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("pwd=%v\n", dir)

	tlsConfig, err := kafka.NewTLSConfig(
		fmt.Sprintf("../../%s", os.Getenv("CLIENT_CERT")),
		fmt.Sprintf("../../%s", os.Getenv("CLIENT_KEY")),
		fmt.Sprintf("../../%s", os.Getenv("SERVER_CERT")))

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

func shutdown() {
	//
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func TestProducer(t *testing.T) {
	fmt.Printf("servers=%v\n", servers)
	producer, err := sarama.NewSyncProducer(servers, config)
	if err != nil {
		t.Fatalf("%s\n", err)
		log.Fatalln(err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			t.Fatalf("%s\n", err)
			log.Fatalln(err)
		}
	}()

	topic := os.Getenv("TOPICS")

	um := UpdateMessage{
		Type:  "Test",
		Phase: "Test",
		Name:  "Test",
		Triple: Triple{
			Subject:   "http://scholars.duke.edu/individual/test00001",
			Predicate: "rdfs:label",
			Object:    "Just a Test"},
	}

	val, err := json.Marshal(um)
	if err != nil {
		t.Fatalf("%s\n", err)
		log.Printf("FAILED to json marshal: %s\n", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(string(val))}
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		t.Fatalf("%s\n", err)
		log.Printf("FAILED to send message: %s\n", err)
	} else {
		log.Printf("> message sent to partition %d at offset %d\n", partition, offset)
	}
}

type ConsumerGroupHandler struct {
	//Callback func()
	//HandleUpdateMessage func(vu.UpdateMessage) //error
	// chan <- updates
}

func (ConsumerGroupHandler) Setup(sess sarama.ConsumerGroupSession) error {
	fmt.Printf("SETUP:%s\n", sess.MemberID())
	return nil
}

func (ConsumerGroupHandler) Cleanup(sess sarama.ConsumerGroupSession) error {
	fmt.Printf("CLEANUP:%s\n", sess.MemberID())
	return nil
}

func (c ConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {

	fmt.Printf("ConsumeClaim()Topic:%s,Partition:%d,Offset:%d,HighWater:%d\n", claim.Topic(),
		claim.Partition(), claim.InitialOffset(),
		claim.HighWaterMarkOffset())

	for msg := range claim.Messages() {
		fmt.Printf("Message topic:%q partition:%d offset:%d\n", msg.Topic, msg.Partition, msg.Offset)

		// used to be consumer.MarkOffset(msg, "metadata")
		sess.MarkMessage(msg, "[testing vivoupdater]")

		um := UpdateMessage{}
		err := json.Unmarshal(msg.Value, &um)

		if err != nil {
			return err
		}
		// NOTE: needs to add to channel
		//updates <- m
		//c.Callback( --> )
		log.Printf("REC:uri=%s\n", um.Triple.Subject)
	}

	return nil
}

func TestConsumer(t *testing.T) {
	fmt.Printf("servers=%v\n", servers)

	topic := strings.Split(os.Getenv("TOPICS"), ",")

	// Start with a client
	/*
		client, err := sarama.NewClient(servers, config)
		if err != nil {
			log.Fatal(err)
		}
		defer func() { _ = client.Close() }()
	*/

	consumer, err := sarama.NewConsumerGroup(servers, "test-group", config)
	if err != nil {
		log.Print(err)
	}
	defer func() { _ = consumer.Close() }()

	// Start a new consumer group
	/*
		group, err := sarama.NewConsumerGroupFromClient("test-group", client)
		if err != nil {
			log.Print(err)
		}
		defer func() { _ = group.Close() }()
	*/

	if err != nil {
		t.Fatalf("%s\n", err)
		log.Fatalln(err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10000*time.Millisecond)

	//handler := ConsumerGroupHandler{Callback: func() { fmt.Println("Hello") }}
	handler := ConsumerGroupHandler{}

	err = consumer.Consume(ctx, topic, handler)
	//err = group.Consume(ctx, topic, handler)

	if err != nil {
		t.Fatalf("%s\n", err)
		log.Fatalln(err)
	}

	//defer consumer.Close()
	defer cancel()
}
