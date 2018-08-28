package vivoupdater

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"io/ioutil"
	"log"
	"os"
)

func NewTLSConfig(clientCertFile, clientKeyFile, caCertFile string) (*tls.Config, error) {
	tlsConfig := tls.Config{}
	// Load client cert
	cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		return &tlsConfig, err
	}
	tlsConfig.Certificates = []tls.Certificate{cert}

	// Load CA cert
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return &tlsConfig, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = caCertPool

	tlsConfig.BuildNameToCertificate()
	return &tlsConfig, err
}

// FIXME: might need all sorts of things
// configurable (acks etc...) maybe
type KafkaSubscriber struct {
	Brokers    []string
	Topics     []string
	ClientCert string
	ClientKey  string
	ServerCert string
	ClientID string
}

func (ks KafkaSubscriber) Subscribe(ctx Context) chan UpdateMessage {
	updates := make(chan UpdateMessage)

	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	brokers := ks.Brokers
	topics := ks.Topics

	ctx.Logger.Printf("%s:%s:%s", ks.ClientCert, ks.ClientKey, ks.ServerCert)
	
	tlsConfig, err := NewTLSConfig(ks.ClientCert, ks.ClientKey, ks.ServerCert)
	//tlsConfig, err := NewTLSConfig("scholars-load-dev.crt.pem", 
	//"scholars-load-dev.key.pem", "kafka-dev-ca.crt.pem")

	if err != nil {
		log.Fatal(err)
	}

	consumerConfig := cluster.NewConfig()
	consumerConfig.ClientID = ks.ClientID // should be config
	consumerConfig.Net.TLS.Enable = true
	consumerConfig.Net.TLS.Config = tlsConfig
	consumerConfig.Consumer.Return.Errors = true

	consumer, err := cluster.NewConsumer(
		brokers,
		"group-id",
		topics,
		consumerConfig)

	if err != nil {
		panic(err)
	}

	go func() {
		// The loop will iterate each time a message is written to the underlying channel
		for msg := range consumer.Messages() {
			var m UpdateMessage
			json.Unmarshal(msg.Value, &m)
			updates <- m
			log.Printf("REC:uri=%s", m.Triple.Subject)

			// Mark the message as processed. The sarama-cluster library will
			// automatically commit these.
			// You can manually commit the offsets using consumer.CommitOffsets()
			consumer.MarkOffset(msg, "required-metadata")
			//break
		}
	}()
	return updates
}
