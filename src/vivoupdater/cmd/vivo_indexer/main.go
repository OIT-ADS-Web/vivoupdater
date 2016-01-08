package main

import (
	"github.com/namsral/flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
	"vivoupdater"
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

func init() {
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
}

func main() {
	go http.ListenAndServe(":8484", nil)
	version := flag.Bool("version", false, "print build id and exit")
	flag.Parse()
	if *version {
		log.Printf("Using build: %s\n", Build)
		os.Exit(0)
	}

	ctx := vivoupdater.Context{
		Notice: vivoupdater.Notification{
			Smtp: notificationSmtp,
			From: notificationFrom,
			To:   []string{notificationEmail}},
		Logger: log.New(os.Stdout, "", log.LstdFlags),
		Quit:   make(chan bool)}

	updates := vivoupdater.UpdateSubscriber{redisUrl, redisChannel, maxRedisAttempts, redisRetryInterval}.Subscribe(ctx)
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
