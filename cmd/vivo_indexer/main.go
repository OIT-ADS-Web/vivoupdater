package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	//"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/namsral/flag"

	"github.com/OIT-ADS-Web/vivoupdater"
	//"github.com/OIT-ADS-Web/vivoupdater/config"
	//"github.com/OIT-ADS-Web/vivoupdater/indexer"
	//"github.com/OIT-ADS-Web/vivoupdater/kafka"
)

func init() {
	flag.StringVar(&vivoupdater.AppEnvironment, "app_environment", "development", "(development|acceptance|production)")
	flag.Var(&vivoupdater.BootstrapFlag, "bootstrap_servers", "comma-separated list of kafka servers")

	// get these from vault
	flag.StringVar(&vivoupdater.VaultEndpoint, "vault_endpoint", "", "vault endpoint")
	flag.StringVar(&vivoupdater.VaultKey, "vault_key", "", "vault key")
	flag.StringVar(&vivoupdater.VaultRoleId, "vault_role_id", "", "vault role id")
	flag.StringVar(&vivoupdater.VaultSecretId, "vault_secret_id", "", "vault secret id")

	// if no vault - get from ssl files ??
	//if len(*vaultRoleID) == 0 && len(*vaultToken) == 0 {
	//  flag.StringVar(&vivoupdater.ClientCert, "client_cert", "", "client ssl cert (*.pem file location)")
	//  flag.StringVar(&vivoupdater.ClientKey, "client_key", "", "client ssl key (*.pem file location)")
	//  flag.StringVar(&vivoupdater.ServerCert, "server_cert", "", "server ssl cert (*.pem file location)")
	// }

	flag.StringVar(&vivoupdater.ClientId, "client_id", "", "client (consumer) id to send to kafka")
	flag.StringVar(&vivoupdater.GroupName, "group_name", "", "client (consumer) group name to send to kafka")

	flag.StringVar(&vivoupdater.MetricsTopic, "metrics_topic", "", "metrics kafka topic")
	flag.StringVar(&vivoupdater.UpdatesTopic, "updates_topic", "", "updates kafka topic")

	// FIXME: do something like this?
	//flag.IntVar(&config.MaxKafkaAttempts, "max_kafak_attempts", 3, "maximum number of consecutive attempts to connect to kafka before exiting")
	//flag.IntVar(&config.KafkaRetryInterval, "kafka_retry_interval", 5, "number of seconds to wait before reconnecting to kafka, reconnects will back off at a rate of num attempts * interval")

	flag.StringVar(&vivoupdater.RedisUrl, "redis_url", "localhost:6379", "host:port of the redis instance")
	flag.StringVar(&vivoupdater.RedisChannel, "redis_channel", "development", "name of the redis channel to subscribe to")
	flag.IntVar(&vivoupdater.MaxRedisAttempts, "max_redis_attempts", 3, "maximum number of consecutive attempts to connect to redis before exiting")
	flag.IntVar(&vivoupdater.RedisRetryInterval, "redis_retry_interval", 5, "number of seconds to wait before reconnecting to redis, reconnects will back off at a rate of num attempts * interval")
	flag.StringVar(&vivoupdater.VivoIndexerUrl, "vivo_indexer_url", "http://localhost:8080/searchService/updateUrisInSearch", "full url of the incremental indexing service")
	flag.StringVar(&vivoupdater.VivoEmail, "vivo_email", "", "email address of vivo user authorized to re-index")
	flag.StringVar(&vivoupdater.VivoPassword, "vivo_password", "", "password for vivo user authorized to re-index")

	flag.StringVar(&vivoupdater.WidgetsIndexerBaseUrl, "widgets_indexer_base_url", "http://localhost:8080/widgets/updates", "base url of the incremental indexing service -  must be expanded in code to differentiate /person vs. /org")

	flag.StringVar(&vivoupdater.WidgetsUser, "widgets_user", "", "email address of vivo user authorized to re-index")
	flag.StringVar(&vivoupdater.WidgetsPassword, "widgets_password", "", "password for vivo user authorized to re-index")
	flag.IntVar(&vivoupdater.BatchSize, "batch_size", 200, "maximum number of uris to send to the indexer at one time")
	flag.IntVar(&vivoupdater.BatchTimeout, "batch_timeout", 10, "maximum number of seconds to wait before sending a partial batch")
	flag.StringVar(&vivoupdater.NotificationSmtp, "notification_smtp", "", "smtp server to use for notifications")
	flag.StringVar(&vivoupdater.NotificationFrom, "notification_from", "", "from address to use for notifications")
	flag.Var(&vivoupdater.NotificationEmail, "notification_email", "email address to use for notifications")

	flag.StringVar(&vivoupdater.LogFile, "log_file", "vivoupdater.log", "rolling log file location")
	flag.IntVar(&vivoupdater.LogMaxSize, "log_max_size", 500, "max size (in mb) of log file")
	flag.IntVar(&vivoupdater.LogMaxBackups, "log_max_backups", 10, "maximum number of old log files to retain")
	flag.IntVar(&vivoupdater.LogMaxAge, "log_max_age", 28, "maximum number of days to keep log file")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// TODO: way to check kafka connection?
	io.WriteString(w, `{"alive": true}`)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", healthCheck)

	profile := flag.Bool("pprof", false, "set this to enable the pprof endpoint")
	if *profile {
		fmt.Println("Enable /debug/pprof/ endpoint for profiling the application.")
		router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	}

	flag.Parse()

	logger := log.New(os.Stdout, "[vivo-updater]", log.LstdFlags)

	logger.SetOutput(&lumberjack.Logger{
		Filename:   vivoupdater.LogFile,
		MaxSize:    vivoupdater.LogMaxSize,
		MaxBackups: vivoupdater.LogMaxBackups,
		MaxAge:     vivoupdater.LogMaxAge,
	})

	vaultConfig := &vivoupdater.VaultConfig{
		Endpoint: vivoupdater.VaultEndpoint,
		AppId:    vivoupdater.VaultKey,
		RoleId:   vivoupdater.VaultRoleId,
		SecretId: vivoupdater.VaultSecretId,
		// e.g. without Token yet
	}

	fmt.Printf("vault-config:%v\n", vaultConfig)
	err := vivoupdater.FetchToken(vaultConfig)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("token=%v\n", vaultConfig.Token)

	kafkaConfig := &vivoupdater.KafkaSubscriber{
		Brokers: vivoupdater.BootstrapFlag,
		Topic:   vivoupdater.UpdatesTopic,
		// NOTE: leaving all cert fields blank
		ClientID:  vivoupdater.ClientId,
		GroupName: vivoupdater.GroupName,
	}
	vivoupdater.GetCertsFromVault(vivoupdater.AppEnvironment,
		vaultConfig, kafkaConfig)

	vivoupdater.SetSubscriber(kafkaConfig)
	// long-lived global producer
	// NOTE: would need to use this to send to other topics
	//producer := vivoupdater.GetProducer()
	err = vivoupdater.SetupProducer(kafkaConfig)

	if err != nil {
		logger.Println("could not establish a kafka async producer")
	}

	// note same subscriber we 'set' above
	consumer := vivoupdater.GetSubscriber()
	context := context.Background()

	for true {
		updates := consumer.Subscribe(context, logger)

		batches := vivoupdater.UriBatcher{
			BatchSize:    vivoupdater.BatchSize,
			BatchTimeout: time.Duration(vivoupdater.BatchTimeout) * time.Second}.Batch(updates)

		vivoIndexer := vivoupdater.VivoIndexer{
			Url:      vivoupdater.VivoIndexerUrl,
			Username: vivoupdater.VivoEmail,
			Password: vivoupdater.VivoPassword}

		widgetsIndexer := vivoupdater.WidgetsIndexer{
			Url:      vivoupdater.WidgetsIndexerBaseUrl,
			Username: vivoupdater.WidgetsUser,
			Password: vivoupdater.WidgetsPassword}

		for b := range batches {
			go vivoupdater.IndexBatch(vivoIndexer, b, logger)
			go vivoupdater.IndexBatch(widgetsIndexer, b, logger)
		}

		// do this or let kafka handle?
		logger.Println("*** Kafka consumer stopped. Wait 5 minutes and try again.")
		time.Sleep(time.Minute * 5)
		//time.Sleep(time.Second * 10)
		logger.Println("*** Start kafka consumer again.")
	}

	//loggedRouter := handlers.CombinedLoggingHandler(os.Stdout, router)
	log.Fatal(http.ListenAndServe(":8484", router))

	//<-ctx.Quit
	//ctx.Logger.Println("Exiting...")
}
