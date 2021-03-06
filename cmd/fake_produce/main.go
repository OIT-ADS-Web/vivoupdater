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
	"github.com/Shopify/sarama"
)

var servers []string
var config *sarama.Config

func init() {
	servers = strings.Split(os.Getenv("BOOTSTRAP_SERVERS"), ",")
	clientId := os.Getenv("CLIENT_ID")
	groupName := os.Getenv("GROUP_NAME")

	appEnv := os.Getenv("APP_ENVIRONMENT")
	vaultApi := os.Getenv("VAULT_ENDPOINT")
	vaultKey := os.Getenv("VAULT_KEY")
	vaultRoleId := os.Getenv("VAULT_ROLE_ID")
	vaultSecretId := os.Getenv("VAULT_SECRET_ID")

	vaultConfig := &vivoupdater.VaultConfig{
		Endpoint: vaultApi,
		AppId:    vaultKey,
		RoleId:   vaultRoleId,
		SecretId: vaultSecretId,
		// e.g. without Token yet
	}
	fmt.Printf("vault-config:%v\n", vaultConfig)

	err := vivoupdater.FetchToken(vaultConfig)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("token=%v\n", vaultConfig.Token)

	kafkaConfig := &vivoupdater.KafkaSubscriber{
		Brokers:   servers,
		ClientID:  clientId,
		GroupName: groupName,
		// e.g. without ClientCert etc...
	}

	// NOTE: some other apps have secrets.properties
	// or secrets.yaml files
	secrets := vivoupdater.SecretsMap(appEnv)
	fmt.Printf("Vault secrets: %s\n", secrets)
	// NOTE: reads values 'into' a struct
	values := &vivoupdater.Secrets{}
	//NOTE: order matters, needs token
	err = vivoupdater.FetchSecrets(vaultConfig, secrets, values)
	if err != nil {
		log.Fatal(err)
	}

	kafkaConfig.ClientCert = values.KafkaClientCert
	kafkaConfig.ClientKey = values.KafkaClientKey
	kafkaConfig.ServerCert = values.KafkaServerCert

	tlsConfig, err := vivoupdater.NewTLSConfig(kafkaConfig)

	if err != nil {
		log.Fatalf("could not read cert file: %v\n", err)
	}
	config = sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	config.Consumer.Return.Errors = true
	//config.Consumer.MaxWaitTime = time.Duration(305000 * time.Millisecond)
	config.ClientID = "rn47" // TODO: config value

	// offset makes a difference
	//config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Net.TLS.Config = tlsConfig
	config.Net.TLS.Enable = true
	config.Version = sarama.V1_0_0_0
}

type ChangeNotifier struct {
	Producer sarama.SyncProducer
	Topic    string
}

func postUri(changeNotifier ChangeNotifier, uri string) {
	um := vivoupdater.UpdateMessage{
		Type:  "update",
		Phase: "destination",
		Name:  "urn:x-arq:UnionGraph",
		Triple: vivoupdater.Triple{
			Subject:   uri,
			Predicate: "rdfs:label",
			Object:    fmt.Sprintf("Specific URI test #%s", uri)},
	}

	val, err := json.Marshal(um)
	if err != nil {
		log.Fatalf("%s\n", err)
		log.Printf("FAILED to json marshal: %s\n", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: changeNotifier.Topic,
		Value: sarama.StringEncoder(string(val))}
	partition, offset, err := changeNotifier.Producer.SendMessage(msg)
	if err != nil {
		log.Fatalf("%s\n", err)
		log.Printf("FAILED to send message: %s\n", err)
	} else {
		log.Printf("subject=%s\n", uri)
		log.Printf("> message sent to partition %d at offset %d\n", partition, offset)
	}
}

func makeFakeUri(changeNotifier ChangeNotifier) {
	num := rand.Intn(9)
	subject := fmt.Sprintf("http://scholars.duke.edu/individual/per000000%d", num)

	postUri(changeNotifier, subject)
}

//https://stackoverflow.com/questions/35809252/check-if-flag-was-provided-in-go
/*
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
*/

func main() {
	number := flag.Int("number", 1, "how many fake uris to post")
	uri := flag.String("uri", "", "a specific uri to post")

	flag.Parse()

	// just making a sync producer - not using 'producer.go'
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

	topic := os.Getenv("UPDATES_TOPIC")

	changeNotifier := ChangeNotifier{
		Producer: producer,
		Topic:    topic,
	}

	if *uri != "" {
		postUri(changeNotifier, *uri)
	} else {
		for i := 1; i <= *number; i++ {
			makeFakeUri(changeNotifier)
		}
	}
}
