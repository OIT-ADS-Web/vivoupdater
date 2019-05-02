package vivoupdater

import (
	"errors"
	"fmt"
	"strings"
)

// stole code from here: https://godoc.org/github.com/namsral/flag
type CSV []string

type KafkaConfig struct {
	BootstrapFlag CSV
	//Topics        CSV
	MetricsTopic string
	UpdatesTopic string
	ClientCert   string
	ClientKey    string
	ServerCert   string
	ClientId     string
	GroupName    string
}

type VaultConfig struct {
	Endpoint string
	RoleId   string
	SecretId string
	AppId    string
	Token    string
}

type Secrets struct {
	KafkaClientKey  string `mapstructure:"kafka_clientKey"`
	KafkaClientCert string `mapstructure:"kafka_clientCert"`
	KafkaServerCert string `mapstructure:"kafka_serverCert"`
}

func SecretsMap(appEnv string) map[string]string {
	secrets := make(map[string]string, 3)
	secrets["kafka.clientKey"] = fmt.Sprintf("apps/scholars/%s/kafka/clientKey", appEnv)
	secrets["kafka.clientCert"] = fmt.Sprintf("apps/scholars/%s/kafka/clientCert", appEnv)
	secrets["kafka.serverCert"] = fmt.Sprintf("apps/scholars/%s/kafka/serverCert", appEnv)
	return secrets
}

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
var AppEnvironment string

// TODO: get rid of notification values entirely?
var NotificationSmtp string
var NotificationFrom string
var NotificationEmail CSV

// logging
var LogFile string
var LogMaxSize int
var LogMaxBackups int
var LogMaxAge int

// kafka
var BootstrapFlag CSV

//var Topics CSV
var UpdatesTopic string
var MetricsTopic string

var ClientCert string
var ClientKey string
var ServerCert string

var ClientId string
var GroupName string

// vault
var VaultEndpoint string
var VaultKey string
var VaultRoleId string
var VaultSecretId string

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
