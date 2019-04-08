package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	vu "github.com/OIT-ADS-Web/vivoupdater"
	"github.com/Shopify/sarama"
)

/*
var Subscriber *KafkaSubscriber

func GetSubscriber() *KafkaSubscriber {
	return Subscriber
}

func SetSubscriber(s *KafkaSubscriber) {
	Subscriber = s
}
*/

// TODO: get from vault
func NewTLSConfig(clientCertFile, clientKeyFile, caCertFile string) (*tls.Config, error) {
	tlsConfig := tls.Config{}
	// Load client cert
	cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		fmt.Println("error reading cert")
		return &tlsConfig, err
	}
	tlsConfig.Certificates = []tls.Certificate{cert}

	// Load CA cert - should get bytes from vault
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		fmt.Println("error reading caCert")
		return &tlsConfig, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = caCertPool

	tlsConfig.BuildNameToCertificate()
	return &tlsConfig, err
}

type ConsumerGroupHandler struct {
	Context vu.Context
	Updates chan vu.UpdateMessage
}

func (ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }

func (ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (c ConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {

	for msg := range claim.Messages() {
		um := vu.UpdateMessage{}
		err := json.Unmarshal(msg.Value, &um)

		if err != nil {
			log.Printf("error: %+v\n", err)
		}
		c.Updates <- um
		// if not succesful ? don't mark offset ??
		// used to be consumer.MarkOffset(msg, "metadata")
		sess.MarkMessage(msg, "")
	}
	return nil
}

type UpdateSubscriber struct {
	Brokers    []string
	Topics     []string
	ClientCert string
	ClientKey  string
	ServerCert string
	ClientID   string
	GroupName  string
}

func startConsumer(ks UpdateSubscriber, handler ConsumerGroupHandler) {
	sarama.Logger = log.New(os.Stdout, "[samara] ", log.LstdFlags)
	tlsConfig, err := NewTLSConfig(ks.ClientCert, ks.ClientKey, ks.ServerCert)

	if err != nil {
		log.Fatal(err)
	}

	consumerConfig := sarama.NewConfig()
	consumerConfig.ClientID = ks.ClientID
	consumerConfig.Version = sarama.V1_0_0_0
	consumerConfig.Net.TLS.Enable = true
	consumerConfig.Net.TLS.Config = tlsConfig
	consumerConfig.Consumer.Return.Errors = true

	consumerConfig.Net.ReadTimeout = (10 * time.Second)
	consumerConfig.Net.DialTimeout = (10 * time.Second)
	consumerConfig.Net.WriteTimeout = (10 * time.Second)

	consumerConfig.Metadata.Retry.Max = 1
	consumerConfig.Metadata.Retry.Backoff = (10 * time.Second)
	consumerConfig.Metadata.RefreshFrequency = (15 * time.Minute)

	// set a max wait time??
	//consumerConfig.Consumer.MaxWaitTime = time.Duration(305000 * time.Millisecond)
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	// go context, not vivoupdater.Context
	context := context.Background()
	client, err := sarama.NewClient(ks.Brokers, consumerConfig)
	if err != nil {
		log.Fatalf("CLIENT ERROR:%v\n", err)
	}
	defer func() { _ = client.Close() }()

	// Start a new consumer group
	group, err := sarama.NewConsumerGroupFromClient(ks.GroupName, client)
	if err != nil {
		log.Printf("GROUP ERROR:%v\n", err)
	}
	defer func() { _ = group.Close() }()

	err = group.Consume(context, ks.Topics, handler)
	if err != nil {
		log.Printf("ERR:%v\n", err)
	}
	defer close(handler.Updates)
}

func (ks UpdateSubscriber) Subscribe(ctx vu.Context) chan vu.UpdateMessage {
	updates := make(chan vu.UpdateMessage)
	handler := ConsumerGroupHandler{Updates: updates, Context: ctx}
	go func() {
		startConsumer(ks, handler)
	}()
	return updates
}
