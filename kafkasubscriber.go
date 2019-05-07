package vivoupdater

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/DCSO/fluxline"
	"github.com/Shopify/sarama"
)

var Subscriber *KafkaSubscriber
var subscribeOnce sync.Once
var setupOnce sync.Once

func GetSubscriber() *KafkaSubscriber {
	return Subscriber
}

func SetSubscriber(s *KafkaSubscriber) {
	Subscriber = s
}

// NOTE: not sure this is necessary
func SetupConsumer(ks *KafkaSubscriber) error {
	subscribeOnce.Do(func() {
		SetSubscriber(ks)
	})
	return nil
}

func NewTLSConfig(ks *KafkaSubscriber) (*tls.Config, error) {
	tlsConfig := tls.Config{}
	// Load client cert
	cert, err := tls.X509KeyPair([]byte(ks.ClientCert), []byte(ks.ClientKey))

	if err != nil {
		fmt.Println("error reading cert")
		return &tlsConfig, err
	}
	tlsConfig.Certificates = []tls.Certificate{cert}

	// Load CA cert - should get bytes from vault
	caCert := []byte(ks.ServerCert)

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

func GetCertsFromVault(env string, config *VaultConfig, kafka *KafkaSubscriber) {
	if len(config.Token) == 0 {
		err := FetchToken(config)
		if err != nil {
			fmt.Printf("Unable to fetch token from vault for role_id and secret_id:\n  %s\n", err)
		}
	}

	secrets := SecretsMap(env)
	// NOTE: reads values 'into' a struct
	var values Secrets
	//NOTE: order matters, needs token
	err := FetchSecrets(config, secrets, &values)
	if err != nil {
		log.Fatal(err)
	}

	kafka.ClientCert = values.KafkaClientCert
	kafka.ClientKey = values.KafkaClientKey
	kafka.ServerCert = values.KafkaServerCert
}

type ConsumerGroupHandler struct {
	Context context.Context
	Logger  *log.Logger
	Updates chan UpdateMessage
}

func (c ConsumerGroupHandler) Setup(sess sarama.ConsumerGroupSession) error {
	c.Logger.Printf("Consumer Setup callback:%s\n", sess.MemberID())
	return nil
}

func (c ConsumerGroupHandler) Cleanup(sess sarama.ConsumerGroupSession) error {
	c.Logger.Printf("Consumer Cleanup callback:%s\n", sess.MemberID())
	return nil
}

// https://github.com/Shopify/sarama/issues/1192
func (c ConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {

	for {
		select {
		case msg := <-claim.Messages():
			var um UpdateMessage
			err := json.Unmarshal(msg.Value, &um)
			if err != nil {
				return err
			}
			c.Logger.Printf("uri consumed: %v\n", um.Triple.Subject)
			c.Updates <- um
			// marking offset
			sess.MarkMessage(msg, "")
		case <-sess.Context().Done():
			err := sess.Context().Err()
			c.Logger.Printf("consumer 'Done': %v\n", err)
			return err
		}
	}
}

type KafkaSubscriber struct {
	Brokers    []string
	Topic      string
	ClientCert string
	ClientKey  string
	ServerCert string
	ClientID   string
	GroupName  string
}

func StartConsumer(ctx context.Context, ks KafkaSubscriber, handler ConsumerGroupHandler) error {
	sarama.Logger = handler.Logger

	tlsConfig, err := NewTLSConfig(&ks)

	if err != nil {
		handler.Logger.Fatal(err)
		return err
	}

	consumerConfig := sarama.NewConfig()
	consumerConfig.ClientID = ks.ClientID
	consumerConfig.Version = sarama.V1_0_0_0
	consumerConfig.Net.TLS.Enable = true
	consumerConfig.Net.TLS.Config = tlsConfig
	consumerConfig.Consumer.Return.Errors = true

	// default is 30 seconds
	consumerConfig.Net.ReadTimeout = (30 * time.Second)
	consumerConfig.Net.DialTimeout = (30 * time.Second)
	consumerConfig.Net.WriteTimeout = (30 * time.Second)

	// not sure a good number of retries
	consumerConfig.Metadata.Retry.Max = 3
	consumerConfig.Metadata.Retry.Backoff = (10 * time.Second)
	// ??? disable
	consumerConfig.Metadata.RefreshFrequency = 0

	// set rebalance timeout?  - default 60ms
	consumerConfig.Consumer.Group.Rebalance.Timeout = time.Duration(10 * time.Second)
	// set a max wait time?? - default 250ms
	//consumerConfig.Consumer.MaxWaitTime = time.Duration(305000 * time.Millisecond)
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	client, err := sarama.NewClient(ks.Brokers, consumerConfig)
	if err != nil {
		// Fatalf calls os.Exit
		handler.Logger.Printf("CLIENT ERROR:%v\n", err)
		return err
	}
	defer func() { _ = client.Close() }()

	// Start a new consumer group
	group, err := sarama.NewConsumerGroupFromClient(ks.GroupName, client)
	if err != nil {
		handler.Logger.Printf("GROUP ERROR:%v\n", err)
		return err
	}
	defer func() { _ = group.Close() }()

	// NOTE: sometimes this gives:
	// CONSUME ERR:kafka server:
	// "A rebalance for the group is in progress. Please re-join the group.
	// Closing Client, Error while closing connection to broker i/o timeout"

	// Track errors
	go func() {
		for err := range group.Errors() {
			fmt.Println("ERROR", err)
		}
	}()

	// see sarama docs of this function about rebalance, retries etc...
	err = group.Consume(ctx, []string{ks.Topic}, handler)
	if err != nil {
		// NOTE: sarama example panics here
		handler.Logger.Printf("CONSUME ERR:%v\n", err)
		return err
	}
	return nil
}

func (ks KafkaSubscriber) Subscribe(ctx context.Context,
	logger *log.Logger, updates chan UpdateMessage) error {
	// NOTE: need to use channel to send to batcher
	handler := ConsumerGroupHandler{Logger: logger, Updates: updates}
	err := StartConsumer(ctx, ks, handler)
	if err != nil {
		logger.Printf("start-consumer error: %v\n", err)
		return err
	}
	return nil
}

// global object
var producer sarama.AsyncProducer

func GetProducer() sarama.AsyncProducer {
	return producer
}

func SetProducer(p sarama.AsyncProducer) {
	producer = p
}

func SetupProducer(ks *KafkaSubscriber) error {
	var setupErr error

	setupOnce.Do(func() {
		tlsConfig, err := NewTLSConfig(ks)
		if err != nil {
			//return err
			setupErr = err
			return
		}
		producerConfig := sarama.NewConfig()
		producerConfig.ClientID = ks.ClientID
		producerConfig.Version = sarama.V1_0_0_0
		producerConfig.Net.TLS.Enable = true
		producerConfig.Net.TLS.Config = tlsConfig

		producer, err := sarama.NewAsyncProducer(ks.Brokers, producerConfig)
		if err != nil {
			//return err
			setupErr = err
			return
		}
		SetProducer(producer)
	})
	return setupErr
	//return nil
}

func Produce(topic string, val string) {
	msg := &sarama.ProducerMessage{Topic: topic, Value: sarama.StringEncoder(val)}
	prod := GetProducer()
	// NOTE: async
	prod.Input() <- msg
}

func FluxLine(measurement string, c interface{}, tags map[string]string) (bytes.Buffer, error) {
	var b bytes.Buffer
	encoder := fluxline.NewEncoder(&b)
	err := encoder.Encode(measurement, c, tags)
	if err != nil {
		return b, err
	}
	return b, nil
}

type IndexMetrics struct {
	Start   time.Time
	End     time.Time
	Uris    []string
	Name    string
	Success bool
}

func SendMetrics(metrics IndexMetrics) {
	rt := (metrics.End.Sub(metrics.Start).Seconds() * 1000.0)

	d := struct {
		Duration float64 `influx:"duration"`
		Count    int64   `influx:"count"`
		Success  bool    `influx:"success"`
	}{
		Duration: rt,
		Count:    int64(len(metrics.Uris)),
		Success:  metrics.Success,
	}

	tags := map[string]string{
		"indexer": metrics.Name,
	}

	line, err := FluxLine("vivoupdater.update", d, tags)
	if err == nil {
		Produce(MetricsTopic, line.String())
	}

}
