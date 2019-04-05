package config

import (
	"errors"
	//"flag"
	"fmt"
	"strings"
	//"github.com/OIT-ADS-Web/vivoupdater/config"
	//"github.com/OIT-ADS-Web/vivoupdater/config"
)

// stole code from here: https://godoc.org/github.com/namsral/flag
type CSV []string

type KafkaConfig struct {
	BootstrapFlag CSV
	Topics        CSV
	ClientCert    string
	ClientKey     string
	ServerCert    string
	ClientId      string
	GroupName     string
}

// redis
type RedisConfig struct {
	RedisUrl           string
	RedisChannel       string
	MaxRedisAttempts   int
	RedisRetryInterval int
}

// vivo
type VivoConfig struct {
	VivoIndexerUrl string
	VivoEmail      string
	VivoPassword   string
}

// widgets
type WidgetsConfig struct {
	WidgetsIndexerBaseUrl string
	WidgetsUser           string
	WidgetsPassword       string
}

// logging
type LoggingConfig struct {
	LogFile       string
	LogMaxSize    int
	LogMaxBackups int
	LogMaxAge     int
}

type ApplicationConfig struct {
	BatchSize         int
	BatchTimeout      int
	NotificationSmtp  string
	NotificationFrom  string
	NotificationEmail string
	Kafka             KafkaConfig
	Widgets           WidgetsConfig
	Vivo              VivoConfig
	Log               LoggingConfig
}

// redis
var RedisUrl string
var RedisChannel string
var MaxRedisAttempts int
var RedisRetryInterval int

// vivo
var VivoIndexerUrl string
var VivoEmail string
var VivoPassword string

// widgets
var WidgetsIndexerBaseUrl string
var WidgetsUser string
var WidgetsPassword string

// misc
var BatchSize int
var BatchTimeout int
var NotificationSmtp string
var NotificationFrom string
var NotificationEmail string

// logging
var LogFile string
var LogMaxSize int
var LogMaxBackups int
var LogMaxAge int

// kafka
var BootstrapFlag CSV
var Topics CSV

var ClientCert string
var ClientKey string
var ServerCert string
var ClientId string
var GroupName string

// kafka
//var bootstrapFlag csv
//var topics csv

//var clientCert string
//var clientKey string
//var serverCert string
//var clientId string
//var groupName string

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (c *CSV) String() string {
	return fmt.Sprint(*c)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (c *CSV) Set(value string) error {
	if len(*c) > 0 {
		return errors.New("flag already set")
	}
	for _, dt := range strings.Split(value, ",") {
		*c = append(*c, dt)
	}
	return nil
}

/*
func Configure() ApplicationConfig {
	var conf ApplicationConfig

	bootstrapFlag := flag.Var("bootstrap_servers", "comma-separated list of kafka servers")
	topics := flag.Var("topics", "comma-separated list of topics")
	clientCert := flag.StringVar("client_cert", "", "client ssl cert (*.pem file location)")
	clientKey := flag.StringVar("client_key", "", "client ssl key (*.pem file location)")
	serverCert := flag.StringVar("server_cert", "", "server ssl cert (*.pem file location)")
	clientId := flag.StringVar("client_id", "", "client (consumer) id to send to kafka")
	groupName := flag.StringVar("group_name", "", "client (consumer) group name to send to kafka")

	flag.Parse()

	kafkaConfig := KafkaConfig{
		BootstrapFlag: bootstrapFlag,
		Topics:        topics,
		ClientCert:    clientCert,
		ClientKey:     clientKey,
		ServerCert:    serverCert,
		ClientId:      clientId,
		GroupName:     groupName,
	}

	conf = ApplicationConfig{
		Kafka: kafkaConfig,
	}
	return conf
}
*/