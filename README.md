# vivoupdater

Vivoupdater subscribes to a Redis channel (or Kafka Topic) containing recently loaded Vivo triples. It will post into both Vivo and Vivo Widgets to trigger selective re-indexing of relevant people and organizations.

##Dependencies

Dependencies are managed with Go Modules

To get started:

     go get github.com/OIT-ADS-Web/vivoupdater
     cd $GOPATH/src/github.com/OIT-ADS-Web/vivoupdater
     go install github.com/OIT-ADS-Web/vivoupdater...

This will create $GOPATH/bin/vivo_indexer

##Configuration

The following environment variables are used to control behavior:

### REDIS... (if Redis enabled)

### REDIS_URL 
  the URL of the redis channel (redis:0000)

### REDIS_CHANNEL

  the name of the redis channel (development.statements)

### BATCH_SIZE

   *default* = 200
   
### BATCH_TIMEOUT

   *default* = 10

### VIVO_INDEXER_URL

   http://localhost:9080/searchService/updateUrisInSearch

### VIVO_EMAIL

   vivo_root@duke.edu

### VIVO_PASSWORD

   <the password>

### WIDGETS_INDEXER_BASE_URL
  
   this will be be appened to with either a /person or /organization specification
   for instance
   
   WIDGETS_INDEXER_BASE_URL=http://localhost:8080/widgets/updates
   
   or for pure testing:
   
   WIDGETS_INDEXER_BASE_URL=http://localhost:3888/updates


### WIDGETS_USER

### WIDGETS_PASSWORD

### NOTIFICATION_SMTP

### NOTIFICATION_FROM

### NOTIFICATION_TO


### KAFKA... (if Kafka enabled)

### BOOTSTRAP_SERVERS

### TOPICS

### CLIENT_CERT

### CLIENT_KEY

### SERVER_CERT

### CLIENT_ID

### GROUP_NAME



