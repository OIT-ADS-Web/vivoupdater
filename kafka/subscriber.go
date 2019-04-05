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

var Subscriber *KafkaSubscriber

func GetSubscriber() *KafkaSubscriber {
	return Subscriber
}

func SetSubscriber(s *KafkaSubscriber) {
	Subscriber = s
}

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

// handler supposed to return error
type ConsumerGroupHandler struct {
	Updates chan vu.UpdateMessage
	Logger  *log.Logger
}

func (ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }

func (ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (c ConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {

	/*
		var msg SyncMessage
		for {
			select {
			case cMsg := <-claim.Messages():
				err := json.Unmarshal(cMsg.Value, &msg)
				if err != nil {
					return err
				}
				// do something
				sess.MarkMessage(cMsg, "")
			case <-sess.Context().Done():
				return nil
			}
		}
	*/

	for msg := range claim.Messages() {
		fmt.Printf("Message topic:%q partition:%d offset:%d\n", msg.Topic, msg.Partition, msg.Offset)

		um := vu.UpdateMessage{}
		err := json.Unmarshal(msg.Value, &um)

		log.Printf("REC:uri=%s\n", um.Triple.Subject)
		c.Logger.Printf("REC:uri=%s\n", um.Triple.Subject)

		if err != nil {
			log.Printf("error: %+v\n", err)
		}
		c.Updates <- um

		// used to be consumer.MarkOffset(msg, "metadata")
		sess.MarkMessage(msg, "")

	}
	return nil
}

type KafkaSubscriber struct {
	Brokers    []string
	Topics     []string
	ClientCert string
	ClientKey  string
	ServerCert string
	ClientID   string
	GroupName  string
}

func (ks KafkaSubscriber) Subscribe(ctx vu.Context) chan vu.UpdateMessage {
	updates := make(chan vu.UpdateMessage)
	sarama.Logger = log.New(os.Stdout, "[samara] ", log.LstdFlags)
	tlsConfig, err := NewTLSConfig(ks.ClientCert, ks.ClientKey, ks.ServerCert)

	if err != nil {
		log.Fatal(err)
	}

	consumerConfig := sarama.NewConfig()
	consumerConfig.ClientID = ks.ClientID // should be config
	consumerConfig.Version = sarama.V1_0_0_0
	consumerConfig.Net.TLS.Enable = true
	consumerConfig.Net.TLS.Config = tlsConfig
	consumerConfig.Consumer.Return.Errors = true

	//consumerConfig.Producer.Retry.Max = 2
	//consumerConfig.Producer.Retry.Backoff = (10 * time.Second)
	//consumerConfig.Producer.Return.Successes = true
	//consumerConfig.Producer.Return.Errors = true
	//consumerConfig.Producer.RequiredAcks = 1
	//consumerConfig.Producer.Timeout = (10 * time.Second)
	consumerConfig.Net.ReadTimeout = (10 * time.Second)
	consumerConfig.Net.DialTimeout = (10 * time.Second)
	consumerConfig.Net.WriteTimeout = (10 * time.Second)

	consumerConfig.Metadata.Retry.Max = 1
	consumerConfig.Metadata.Retry.Backoff = (10 * time.Second)
	consumerConfig.Metadata.RefreshFrequency = (15 * time.Minute)

	//consumerConfig.Consumer.MaxWaitTime = time.Duration(305000 * time.Millisecond)
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	// Start with a client
	client, err := sarama.NewClient(ks.Brokers, consumerConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	// Start a new consumer group
	group, err := sarama.NewConsumerGroupFromClient(ks.GroupName, client)
	if err != nil {
		log.Print(err)
	}
	defer func() { _ = group.Close() }()

	// go context, not our own
	context := context.Background()

	// is go wrapper necessary here?
	go func() {
		handler := ConsumerGroupHandler{Updates: updates,
			Logger: ctx.Logger}
		// NOTE: was in loop
		//err := group.Consume(ctx, []string{ks.ChangesTopic}, handler)

		err = group.Consume(context, ks.Topics, handler)
		if err != nil {
			log.Print(err)
		}

	}()
	return updates
}
