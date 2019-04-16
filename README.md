# vivoupdater

Vivoupdater subscribes to a Kafka Topic containing recently loaded Vivo triples. It will post into both Vivo and Vivo Widgets to trigger selective re-indexing of relevant people and organizations.

## Dependencies

Dependencies are managed with Go Modules

To install:

```bash
go get github.com/OIT-ADS-Web/vivoupdater

cd $GOPATH/src/github.com/OIT-ADS-Web/vivoupdater
     
go install github.com/OIT-ADS-Web/vivoupdater...
```

This will create $GOPATH/bin/vivo_indexer and $GOPATH/bin/fake_produce.

## Configuration

The following environment variables are used to control behavior:

### APP_ENVIRONMENT

The environment the code is running in (e.g. development|acceptance|production).  Also determines where Vault looks for keys (see next)

### VAULT_ENDPOINT

The endpoint api of your vault installation

### VAULT_KEY

Using the app_role auth method - this is the key value

### VAULT_ROLE_ID

The vault role id

### VAULT_SECRET_ID

The vault secret id value

### BATCH_SIZE

*default* = 200
   
### BATCH_TIMEOUT

*default* = 10

### VIVO_INDEXER_URL

Where the indexer should post vivo updates

For example:

http://localhost:9080/searchService/updateUrisInSearch

### VIVO_EMAIL

The vivo instance admin email, for example:

vivo_root@duke.edu

### VIVO_PASSWORD

The vivo instance admin password

### WIDGETS_INDEXER_BASE_URL

Where the indexer should post widgets updates

this will be be appened to with either a /person or /organization specification
   
for example:
   
http://localhost:8080/widgets/updates
   

### WIDGETS_USER

widgets instance user

### WIDGETS_PASSWORD

widgets instance password

### NOTIFICATION_SMTP

*optional* - smtp server to send emails when service goes
down

### NOTIFICATION_FROM

*optional* - (unless NOTIFICATION_SMTP is present) - the [FROM] value

### NOTIFICATION_TO

*optional* - (unless NOTIFICATION_SMTP is present) - the [TO] value
 
### BOOTSTRAP_SERVERS

The Kafka list of servers (as comma separated list)

### UPDATES_TOPIC

The Kafka topic that has the updates - just one at this
point

### METRICS_TOPIC

The kafka topic to send metrics

### CLIENT_ID

A kafka client name

### GROUP_NAME

A kafka group name

## Debugging

I've found pprof helpful.  You can send in the flag `pprof` like this:

> vivo_indexer -pprof=true

and it will enable (see go pprof):

For example (with Graphviz installed):

> go tool pprof -png http://localhost:8484/debug/pprof/heap > out.png
