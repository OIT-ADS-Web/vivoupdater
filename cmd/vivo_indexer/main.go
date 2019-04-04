package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/OIT-ADS-Web/vivoupdater"
	"github.com/namsral/flag"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Build string

// redis
var redisUrl string
var redisChannel string
var maxRedisAttempts int
var redisRetryInterval int

// vivo
var vivoIndexerUrl string
var vivoEmail string
var vivoPassword string

// widgets
var widgetsIndexerBaseUrl string
var widgetsUser string
var widgetsPassword string

// misc
var batchSize int
var batchTimeout int
var notificationSmtp string
var notificationFrom string
var notificationEmail string

// logging
var logFile string
var logMaxSize int
var logMaxBackups int
var logMaxAge int

// stole code from here: https://godoc.org/github.com/namsral/flag
type csv []string

var bootstrapFlag csv
var topics csv

var clientCert string
var clientKey string
var serverCert string
var clientId string
var groupName string

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (c *csv) String() string {
	return fmt.Sprint(*c)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (c *csv) Set(value string) error {
	if len(*c) > 0 {
		return errors.New("flag already set")
	}
	for _, dt := range strings.Split(value, ",") {
		*c = append(*c, dt)
	}
	return nil
}

func init() {
	flag.Var(&bootstrapFlag, "bootstrap_servers", "comma-separated list of kafka servers")
	flag.Var(&topics, "topics", "comma-separated list of topics")
	flag.StringVar(&clientCert, "client_cert", "", "client ssl cert (*.pem file location)")
	flag.StringVar(&clientKey, "client_key", "", "client ssl key (*.pem file location)")
	flag.StringVar(&serverCert, "server_cert", "", "server ssl cert (*.pem file location)")
	flag.StringVar(&clientId, "client_id", "", "client (consumer) id to send to kafka")
	flag.StringVar(&groupName, "group_name", "", "client (consumer) group name to send to kafka")

	flag.StringVar(&redisUrl, "redis_url", "localhost:6379", "host:port of the redis instance")
	flag.StringVar(&redisChannel, "redis_channel", "development", "name of the redis channel to subscribe to")
	flag.IntVar(&maxRedisAttempts, "max_redis_attempts", 3, "maximum number of consecutive attempts to connect to redis before exiting")
	flag.IntVar(&redisRetryInterval, "redis_retry_interval", 5, "number of seconds to wait before reconnecting to redis, reconnects will back off at a rate of num attempts * interval")
	flag.StringVar(&vivoIndexerUrl, "vivo_indexer_url", "http://localhost:8080/searchService/updateUrisInSearch", "full url of the incremental indexing service")
	flag.StringVar(&vivoEmail, "vivo_email", "", "email address of vivo user authorized to re-index")
	flag.StringVar(&vivoPassword, "vivo_password", "", "password for vivo user authorized to re-index")

	flag.StringVar(&widgetsIndexerBaseUrl, "widgets_indexer_base_url", "http://localhost:8080/widgets/updates", "base url of the incremental indexing service -  must be expanded in code to differentiate /person vs. /org")

	flag.StringVar(&widgetsUser, "widgets_user", "", "email address of vivo user authorized to re-index")
	flag.StringVar(&widgetsPassword, "widgets_password", "", "password for vivo user authorized to re-index")
	flag.IntVar(&batchSize, "batch_size", 200, "maximum number of uris to send to the indexer at one time")
	flag.IntVar(&batchTimeout, "batch_timeout", 10, "maximum number of seconds to wait before sending a partial batch")
	flag.StringVar(&notificationSmtp, "notification_smtp", "", "smtp server to use for notifications")
	flag.StringVar(&notificationFrom, "notification_from", "", "from address to use for notifications")
	flag.StringVar(&notificationEmail, "notification_email", "", "email address to use for notifications")

	flag.StringVar(&logFile, "log_file", "vivoupdater.log", "rolling log file location")
	flag.IntVar(&logMaxSize, "log_max_size", 500, "max size (in mb) of log file")
	flag.IntVar(&logMaxBackups, "log_max_backups", 10, "maximum number of old log files to retain")
	flag.IntVar(&logMaxAge, "log_max_age", 28, "maximum number of days to keep log file")
}

func main() {
	go http.ListenAndServe(":8484", nil)
	version := flag.Bool("version", false, "print build id and exit")
	flag.Parse()
	if *version {
		log.Printf("Using build: %s\n", Build)
		os.Exit(0)
	}

	var log = log.New(os.Stdout, "", log.LstdFlags)

	log.SetOutput(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    logMaxSize,
		MaxBackups: logMaxBackups,
		MaxAge:     logMaxAge,
	})

	ctx := vivoupdater.Context{
		Notice: vivoupdater.Notification{
			Smtp: notificationSmtp,
			From: notificationFrom,
			To:   []string{notificationEmail}},
		Logger: log,
		Quit:   make(chan bool)}

	// something like this:
	// how to get values from vault ---?
	/*
			type KafkaSubscriber struct {
			Brokers    []string
			Topics     []string
			ClientCert string
			ClientKey  string
			ServerCert string
			ClientID   string
			GroupName  string
		}
	*/
	updates := vivoupdater.KafkaSubscriber{Brokers: bootstrapFlag,
		Topics:     topics,
		ClientCert: clientCert,
		ClientKey:  clientKey,
		ServerCert: serverCert,
		ClientID:   clientId,
		GroupName:  groupName}.Subscribe(ctx)
	//updates := vivoupdater.UpdateSubscriber{redisUrl, redisChannel, maxRedisAttempts, redisRetryInterval}.Subscribe(ctx)
	batches := vivoupdater.UriBatcher{batchSize, time.Duration(batchTimeout) * time.Second}.Batch(ctx, updates)

	vivoIndexer := vivoupdater.VivoIndexer{vivoIndexerUrl, vivoEmail, vivoPassword}
	widgetsIndexer := vivoupdater.WidgetsIndexer{widgetsIndexerBaseUrl, widgetsUser, widgetsPassword}

	for b := range batches {
		go vivoupdater.IndexBatch(ctx, vivoIndexer, b)
		go vivoupdater.IndexBatch(ctx, widgetsIndexer, b)
	}

	<-ctx.Quit
	ctx.Logger.Println("Exiting...")
}
