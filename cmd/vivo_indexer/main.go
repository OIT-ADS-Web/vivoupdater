package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	//"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/OIT-ADS-Web/vivoupdater"
	"github.com/namsral/flag"

	"github.com/OIT-ADS-Web/vivoupdater/config"
)

func init() {
	flag.Var(&config.BootstrapFlag, "bootstrap_servers", "comma-separated list of kafka servers")
	flag.Var(&config.Topics, "topics", "comma-separated list of topics")
	flag.StringVar(&config.ClientCert, "client_cert", "", "client ssl cert (*.pem file location)")
	flag.StringVar(&config.ClientKey, "client_key", "", "client ssl key (*.pem file location)")
	flag.StringVar(&config.ServerCert, "server_cert", "", "server ssl cert (*.pem file location)")
	flag.StringVar(&config.ClientId, "client_id", "", "client (consumer) id to send to kafka")
	flag.StringVar(&config.GroupName, "group_name", "", "client (consumer) group name to send to kafka")

	flag.StringVar(&config.RedisUrl, "redis_url", "localhost:6379", "host:port of the redis instance")
	flag.StringVar(&config.RedisChannel, "redis_channel", "development", "name of the redis channel to subscribe to")
	flag.IntVar(&config.MaxRedisAttempts, "max_redis_attempts", 3, "maximum number of consecutive attempts to connect to redis before exiting")
	flag.IntVar(&config.RedisRetryInterval, "redis_retry_interval", 5, "number of seconds to wait before reconnecting to redis, reconnects will back off at a rate of num attempts * interval")
	flag.StringVar(&config.VivoIndexerUrl, "vivo_indexer_url", "http://localhost:8080/searchService/updateUrisInSearch", "full url of the incremental indexing service")
	flag.StringVar(&config.VivoEmail, "vivo_email", "", "email address of vivo user authorized to re-index")
	flag.StringVar(&config.VivoPassword, "vivo_password", "", "password for vivo user authorized to re-index")

	flag.StringVar(&config.WidgetsIndexerBaseUrl, "widgets_indexer_base_url", "http://localhost:8080/widgets/updates", "base url of the incremental indexing service -  must be expanded in code to differentiate /person vs. /org")

	flag.StringVar(&config.WidgetsUser, "widgets_user", "", "email address of vivo user authorized to re-index")
	flag.StringVar(&config.WidgetsPassword, "widgets_password", "", "password for vivo user authorized to re-index")
	flag.IntVar(&config.BatchSize, "batch_size", 200, "maximum number of uris to send to the indexer at one time")
	flag.IntVar(&config.BatchTimeout, "batch_timeout", 10, "maximum number of seconds to wait before sending a partial batch")
	flag.StringVar(&config.NotificationSmtp, "notification_smtp", "", "smtp server to use for notifications")
	flag.StringVar(&config.NotificationFrom, "notification_from", "", "from address to use for notifications")
	flag.StringVar(&config.NotificationEmail, "notification_email", "", "email address to use for notifications")

	flag.StringVar(&config.LogFile, "log_file", "vivoupdater.log", "rolling log file location")
	flag.IntVar(&config.LogMaxSize, "log_max_size", 500, "max size (in mb) of log file")
	flag.IntVar(&config.LogMaxBackups, "log_max_backups", 10, "maximum number of old log files to retain")
	flag.IntVar(&config.LogMaxAge, "log_max_age", 28, "maximum number of days to keep log file")
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

	//go http.ListenAndServe(":8484", nil)
	var log = log.New(os.Stdout, "[vivo-updater]", log.LstdFlags)

	/*
		log.SetOutput(&lumberjack.Logger{
			Filename:   config.LogFile,
			MaxSize:    config.LogMaxSize,
			MaxBackups: config.LogMaxBackups,
			MaxAge:     config.LogMaxAge,
		})
	*/

	log.Println("started...")
	ctx := vivoupdater.Context{
		Notice: vivoupdater.Notification{
			Smtp: config.NotificationSmtp,
			From: config.NotificationFrom,
			To:   []string{config.NotificationEmail}},
		Logger: log,
		Quit:   make(chan bool)}

	updates := vivoupdater.KafkaSubscriber{
		Brokers:    config.BootstrapFlag,
		Topics:     config.Topics,
		ClientCert: config.ClientCert,
		ClientKey:  config.ClientKey,
		ServerCert: config.ServerCert,
		ClientID:   config.ClientId,
		GroupName:  config.GroupName}.Subscribe(ctx)

	batches := vivoupdater.UriBatcher{
		BatchSize:    config.BatchSize,
		BatchTimeout: time.Duration(config.BatchTimeout) * time.Second}.Batch(ctx, updates)

	vivoIndexer := vivoupdater.VivoIndexer{
		Url:      config.VivoIndexerUrl,
		Username: config.VivoEmail,
		Password: config.VivoPassword}

	widgetsIndexer := vivoupdater.WidgetsIndexer{
		Url:      config.WidgetsIndexerBaseUrl,
		Username: config.WidgetsUser,
		Password: config.WidgetsPassword}

	for b := range batches {
		go vivoupdater.IndexBatch(ctx, vivoIndexer, b)
		go vivoupdater.IndexBatch(ctx, widgetsIndexer, b)
	}

	//loggedRouter := handlers.CombinedLoggingHandler(os.Stdout, router)

	log.Fatal(http.ListenAndServe(":8484", router))

	<-ctx.Quit
	ctx.Logger.Println("Exiting...")
}
