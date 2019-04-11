package vivoupdater

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/DCSO/fluxline"
	"github.com/Shopify/sarama"
)

var Subscriber *KafkaSubscriber

func GetSubscriber() *KafkaSubscriber {
	return Subscriber
}

func SetSubscriber(s *KafkaSubscriber) {
	Subscriber = s
}

func SetupConsumer(ks *KafkaSubscriber) error {
	SetSubscriber(ks)
	return nil
}

func NewTLSConfig(ks *KafkaSubscriber) (*tls.Config, error) {
	tlsConfig := tls.Config{}
	// Load client cert
	// NOTE: this will need to change to read text
	//cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	cert, err := tls.X509KeyPair([]byte(ks.ClientCert), []byte(ks.ClientKey))

	if err != nil {
		fmt.Println("error reading cert")
		return &tlsConfig, err
	}
	tlsConfig.Certificates = []tls.Certificate{cert}

	// Load CA cert - should get bytes from vault
	//caCert, err := ioutil.ReadFile(caCertFile)
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
	//Brokers:   servers,
	//ClientID:  vivoupdater.ClientId,
	//GroupName: vivoupdater.GroupName,

	secrets := SecretsMap(env)
	//fmt.Printf("Vault secrets: %s\n", secrets)
	// NOTE: reads values 'into' a struct
	var values Secrets
	//NOTE: order matters, needs token

	// maybe this could be a copy -> into -> method squash
	//kafkaConfig.ClientCert = values.Kafka.ClientCert
	//kafkaConfig.ClientKey = values.Kafka.ClientKey
	//kafkaConfig.ServerCert = values.Kafka.ServerCert

	/*
		path := fmt.Sprintf("secret/apps/scholars/%s/kafka", env)
		//fmt.Printf("Vault path: %s\n", path)
		var kafka *KafkaSubscriber
		err := FetchValues(config, path, kafka)
	*/
	//var kafka *KafkaSubscriber
	err := FetchSecrets(config, secrets, &values)
	if err != nil {
		log.Fatal(err)
	}

	//kafka.Brokers = BootstrapFlag
	//kafka.ClientID = ClientId
	//kafka.GroupName = GroupName
	kafka.ClientCert = values.KafkaClientCert
	kafka.ClientKey = values.KafkaClientKey
	kafka.ServerCert = values.KafkaServerCert

	/*
		if err != nil {
			fmt.Println("Error fetching values from vault. Waiting 30 seconds before stopping.")
			fmt.Println(err)
			time.Sleep(time.Second * 30)
		}
	*/

	//SetupSubscriber(kafka)
	//SetupProducer(kafka)
}

type ConsumerGroupHandler struct {
	Context context.Context
	Logger  *log.Logger
	Updates chan UpdateMessage
	Cancel  context.CancelFunc
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

	um := UpdateMessage{}

	for {
		select {
		case msg := <-claim.Messages():
			err := json.Unmarshal(msg.Value, &um)
			if err != nil {
				return err
			}
			c.Logger.Printf("uri received: %v\n", um.Triple.Subject)
			c.Updates <- um
			sess.MarkMessage(msg, "")
		case <-sess.Context().Done():
			err := sess.Context().Err()
			notifier := GetNotifier()
			notifier.DoSend("Kafka Subscriber shutdown via sarama Context", err)
			return nil
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

	consumerConfig.Net.ReadTimeout = (10 * time.Second)
	consumerConfig.Net.DialTimeout = (10 * time.Second)
	consumerConfig.Net.WriteTimeout = (10 * time.Second)

	consumerConfig.Metadata.Retry.Max = 1
	consumerConfig.Metadata.Retry.Backoff = (10 * time.Second)
	consumerConfig.Metadata.RefreshFrequency = (15 * time.Minute)

	// set a max wait time??
	//consumerConfig.Consumer.MaxWaitTime = time.Duration(305000 * time.Millisecond)
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	//consumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	client, err := sarama.NewClient(ks.Brokers, consumerConfig)
	if err != nil {
		handler.Logger.Fatalf("CLIENT ERROR:%v\n", err)
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
	err = group.Consume(ctx, []string{ks.Topic}, handler)
	if err != nil {
		// NOTE: sarama example panics here
		handler.Logger.Printf("CONSUME ERR:%v\n", err)
		return err
	}
	return nil
}

//https://dave.cheney.net/tag/logging
func (ks KafkaSubscriber) Subscribe(ctx context.Context,
	logger *log.Logger) chan UpdateMessage {
	// NOTE: need to use channel to send to batcher
	updates := make(chan UpdateMessage)

	handler := ConsumerGroupHandler{Logger: logger, Updates: updates}
	go func() {
		// not sure if this will catch consumer rebalance error or not
		err := StartConsumer(ctx, ks, handler)
		if err != nil {
			logger.Printf("start-consumer error: %v\n", err)
			panic(err)
		}
	}()
	return updates
}

var producer sarama.AsyncProducer

func GetProducer() sarama.AsyncProducer {
	return producer
}

func SetProducer(p sarama.AsyncProducer) {
	producer = p
}

func SetupProducer(ks *KafkaSubscriber) error {
	tlsConfig, err := NewTLSConfig(ks)
	if err != nil {
		return err
	}
	producerConfig := sarama.NewConfig()
	producerConfig.ClientID = ks.ClientID
	producerConfig.Version = sarama.V1_0_0_0
	producerConfig.Net.TLS.Enable = true
	producerConfig.Net.TLS.Config = tlsConfig

	producer, err := sarama.NewAsyncProducer(ks.Brokers, producerConfig)
	if err != nil {
		return err
	}

	SetProducer(producer)
	return nil
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
	Start time.Time
	End   time.Time
	Uris  []string
	Name  string
}

func SendMetrics(metrics IndexMetrics, logger *log.Logger) {
	rt := (metrics.End.Sub(metrics.Start).Seconds() * 1000.0)

	logger.Println("...sending some metrics")
	d := struct {
		Duration float64 `influx:"duration"`
		Count    int64   `influx:"count"`
	}{
		Duration: rt,
		Count:    int64(len(metrics.Uris)),
	}

	tags := map[string]string{
		"indexer": metrics.Name,
	}

	line, err := FluxLine("vivoupdater.update", d, tags)
	if err == nil {
		Produce(MetricsTopic, line.String())
	}

}
