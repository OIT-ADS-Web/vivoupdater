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

//type Message struct {
//	URI  string `json:"uri"`
//	Type string `json:"type"`
//}

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

type KafkaSubscriber struct {
	Brokers []string
	Topics  []string
}

func (ks KafkaSubscriber) Subscribe(ctx Context) chan UpdateMessage {
	updates := make(chan UpdateMessage)

	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	brokers := ks.Brokers
	//brokers := []string{"kafka-dev-01.oit.duke.edu:9093",
	//		"kafka-dev-02.oit.duke.edu:9093", "kafka-dev-03.oit.duke.edu:9093"}
	//topics := []string{"scholars-resources-changed"}
	topics := ks.Topics

	tlsConfig, err := NewTLSConfig("scholars-load-dev.crt.pem",
		"scholars-load-dev.key.pem",
		"kafka-dev-ca.crt.pem")
	if err != nil {
		log.Fatal(err)
	}

	consumerConfig := cluster.NewConfig()
	consumerConfig.ClientID = "rn47" // should be config
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

	//ReceiveLoop:
		// The loop will iterate each time a message is written to the underlying channel
		for msg := range consumer.Messages() {
			// Now we can access the individual fields of the message and react
			// based on msg.Topic
			switch msg.Topic {
			case "scholars-resources-changed":
				// Do everything we need for this topic
				//n := bytes.IndexByte(msg.Value, 0)
				var m UpdateMessage
				json.Unmarshal(msg.Value, &m)
				updates <- m
				//s := string(msg.Value[:])
				log.Printf("%s", m)
				log.Printf("uri=%s;type=%s", m.Triple.Subject, m.Type)

				// Mark the message as processed. The sarama-cluster library will
				// automatically commit these.
				// You can manually commit the offsets using consumer.CommitOffsets()
				consumer.MarkOffset(msg, "required-metadata")
				break //ReceiveLoop
				// ...
			}
		}
	}()

	/*
		go func() {
			failedAttempts := 0
			for {
				ctx.Logger.Println("Connecting to Redis at " + us.RedisUrl)
				c, err := redis.Dial("tcp", us.RedisUrl)
					if err != nil {
						failedAttempts += 1
						if failedAttempts > us.MaxConnectAttempts {
							close(updates)
							ctx.handleError("Max redis connection attempts exceeded", err, true)
							break
						}
						ctx.handleError("Could not connect to Redis", err, false)
						time.Sleep(time.Duration(us.RetryInterval*failedAttempts) * time.Second)
						continue
					}
					failedAttempts = 0
					psc := redis.PubSubConn{c}
					psc.Subscribe(us.RedisChannel)
					ctx.Logger.Println("Subscribed to channel " + us.RedisChannel)

				ReceiveLoop:
					for {
						switch v := psc.Receive().(type) {
						case redis.Message:
							var m UpdateMessage
							json.Unmarshal(v.Data, &m)
							updates <- m
						case error:
							ctx.handleError("Lost connection to Redis", v, false)
							psc.Close()
							time.Sleep(time.Duration(us.RetryInterval) * time.Second)
							failedAttempts += 1
							break ReceiveLoop
						}
					}
				}
			}()
	*/

	return updates
}
