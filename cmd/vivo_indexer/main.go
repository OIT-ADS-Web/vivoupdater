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

	"github.com/gorilla/mux"

	"github.com/namsral/flag"

	"github.com/OIT-ADS-Web/vivoupdater"
)

func init() {
	flag.StringVar(&vivoupdater.AppEnvironment, "app_environment", "development", "(development|acceptance|production)")
	flag.Var(&vivoupdater.BootstrapFlag, "bootstrap_servers", "comma-separated list of kafka servers")

	// get these from vault
	flag.StringVar(&vivoupdater.VaultEndpoint, "vault_endpoint", "", "vault endpoint")
	flag.StringVar(&vivoupdater.VaultKey, "vault_key", "", "vault key")
	flag.StringVar(&vivoupdater.VaultRoleId, "vault_role_id", "", "vault role id")
	flag.StringVar(&vivoupdater.VaultSecretId, "vault_secret_id", "", "vault secret id")

	flag.StringVar(&vivoupdater.ClientId, "client_id", "", "client (consumer) id to send to kafka")
	flag.StringVar(&vivoupdater.GroupName, "group_name", "", "client (consumer) group name to send to kafka")

	flag.StringVar(&vivoupdater.MetricsTopic, "metrics_topic", "", "metrics kafka topic")
	flag.StringVar(&vivoupdater.UpdatesTopic, "updates_topic", "", "updates kafka topic")

	// TODO: add something like this?
	//flag.IntVar(&config.MaxKafkaAttempts, "max_kafka_attempts", 3, "maximum number of consecutive attempts to connect to kafka before exiting")
	//flag.IntVar(&config.KafkaRetryInterval, "kafka_retry_interval", 5, "number of seconds to wait before reconnecting to kafka, reconnects will back off at a rate of num attempts * interval")

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
	srv := &http.Server{
		Addr: ":8484",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router, // Pass our instance of gorilla/mux in.
	}

	profile := flag.Bool("pprof", false, "set this to enable the pprof endpoint")

	flag.Parse()
	if *profile {
		fmt.Println("Enable /debug/pprof/ endpoint for profiling the application.")
		router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	}

	logger := log.New(os.Stdout, "[vivo-updater]", log.LstdFlags)
	ctx := context.Background()
	cancellable, cancel := context.WithCancel(ctx)
	// TODO: is this the correct usage of context cancel?
	// https://dave.cheney.net/2017/08/20/context-isnt-for-cancellation
	// but
	// supposedly catching 'consumer rebalance' error involves
	// checking context.Done()
	// https://github.com/Shopify/sarama/issues/1192

	// TODO: not sure this is ever actually called
	defer func() {
		cancel() // unnecessary call?
		if err := srv.Shutdown(ctx); err != nil {
			logger.Fatalf("could not shutdown: %v", err)
		}
		logger.Println("shutting down")
		os.Exit(1)
	}()

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Fatalf("could not start server: %v", err)
			cancel() // unnecessary call?
			os.Exit(1)
		}
	}()

	vaultConfig := &vivoupdater.VaultConfig{
		Endpoint: vivoupdater.VaultEndpoint,
		AppId:    vivoupdater.VaultKey,
		RoleId:   vivoupdater.VaultRoleId,
		SecretId: vivoupdater.VaultSecretId,
		// e.g. without Token yet
	}

	err := vivoupdater.FetchToken(vaultConfig)
	if err != nil {
		log.Fatal(err)
	}

	kafkaConfig := &vivoupdater.KafkaSubscriber{
		Brokers: vivoupdater.BootstrapFlag,
		Topic:   vivoupdater.UpdatesTopic,
		// e.g. all cert fields blank
		ClientID:  vivoupdater.ClientId,
		GroupName: vivoupdater.GroupName,
	}
	vivoupdater.GetCertsFromVault(vivoupdater.AppEnvironment,
		vaultConfig, kafkaConfig)

	vivoupdater.SetSubscriber(kafkaConfig)
	// long-lived global producer
	err = vivoupdater.SetupProducer(kafkaConfig)

	if err != nil {
		logger.Println("could not establish a kafka async producer")
	}

	// note same subscriber we 'set' above
	consumer := vivoupdater.GetSubscriber()

	updates := make(chan vivoupdater.UpdateMessage)

	batches := vivoupdater.UriBatcher{
		BatchSize:    vivoupdater.BatchSize,
		BatchTimeout: time.Duration(vivoupdater.BatchTimeout) * time.Second}.Batch(cancellable, updates)

	vivoIndexer := vivoupdater.VivoIndexer{
		Url:      vivoupdater.VivoIndexerUrl,
		Username: vivoupdater.VivoEmail,
		Password: vivoupdater.VivoPassword}

	widgetsIndexer := vivoupdater.WidgetsIndexer{
		Url:      vivoupdater.WidgetsIndexerBaseUrl,
		Username: vivoupdater.WidgetsUser,
		Password: vivoupdater.WidgetsPassword}

	go func() {
		err := consumer.Subscribe(cancellable, logger, updates)
		if err != nil {
			// how to 're-start' here? e.g. capture rebalance
			logger.Printf("consumer subscribe error:%v\n", err)
			cancel() // unnecessary call?
			os.Exit(1)
		}
	}()

	// main for-loop of application
IndexerLoop:

	for {
		select {
		case b := <-batches:
			go vivoupdater.IndexBatch(vivoIndexer, b, logger)
			go vivoupdater.IndexBatch(widgetsIndexer, b, logger)
		case <-ctx.Done():
			cancel() // unnecessary call?
			logger.Printf("indexer loop closed because %v\n", ctx.Err())
			// NOTE: trying to avoid infinite loop in updates.go
			// (e.g. sending Notifier emails over and over)
			// would sleep? or some kind of backoff timeout be better?
			os.Exit(1)
			// I thought breaking here would effectively exit
			// but it did not seem to do that
			break IndexerLoop
		}
	}

}
