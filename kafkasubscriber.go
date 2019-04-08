package vivoupdater

/*
var Subscriber *KafkaSubscriber

func GetSubscriber() *KafkaSubscriber {
	return Subscriber
}

func SetSubscriber(s *KafkaSubscriber) {
	Subscriber = s
}

// TODO: get from vault
func NewTLSConfig(clientCertFile, clientKeyFile, caCertFile string) (*tls.Config, error) {
	tlsConfig := tls.Config{}
	// Load client cert
	cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		return &tlsConfig, err
	}
	tlsConfig.Certificates = []tls.Certificate{cert}

	// Load CA cert - should get bytes from vault
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return &tlsConfig, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = caCertPool

	tlsConfig.BuildNameToCertificate()
	return &tlsConfig, err
}


// handler supposed to return error
type consumerGroupHandler struct {
	HandleUpdateMessage func(UpdateMessage) //error
}

func (consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }

func (consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

// must return 'error'
func (h consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	// go func?
	// this is a channel?
	// Messages() <-chan *ConsumerMessage
	for msg := range claim.Messages() {
		fmt.Printf("Message topic:%q partition:%d offset:%d\n", msg.Topic, msg.Partition, msg.Offset)
		// used to be consumer.MarkOffset(msg, "metadata")
		sess.MarkMessage(msg, "")

		um := UpdateMessage{}
		err := json.Unmarshal(msg.Value, &um)

		//json.Unmarshal(msg.Value, &m)
		// NOTE: needs to add to channel
		//updates <- m
		log.Printf("REC:uri=%s\n", um.Triple.Subject)

		//updates <- um
		// TODO: way to apply backpressure here - or somewhere?
		// Mark the message as processed. The sarama-cluster library will
		// automatically commit these.
		// You can manually commit the offsets using consumer.CommitOffsets()
		//consumer.MarkOffset(msg, "required-metadata")

		if err != nil {
			log.Printf("error: %+v\n", err)
		}
		if h.HandleUpdateMessage != nil {
			//e.g. add to channel updates <- um
			h.HandleUpdateMessage(um)
		}
	}
	return nil
}

type KafkaSubscriber struct {
	Brokers    []string
	Topics     []string
	ClientCert string
	ClientKey  string
	ServerCert string
	ClientID   string
	GroupName  string
}

// callback ?
//ctx Context) chan UpdateMessage
func (ks KafkaSubscriber) Subscribe(ctx Context) chan UpdateMessage {
	updates := make(chan UpdateMessage)
	//func (ks *KafkaSubscriber) Subscribe(h func(UpdateMessage) error) {
	sarama.Logger = log.New(os.Stdout, "[samara] ", log.LstdFlags)
	tlsConfig, err := NewTLSConfig(ks.ClientCert, ks.ClientKey, ks.ServerCert)

	if err != nil {
		log.Fatal(err)
	}

	consumerConfig := sarama.NewConfig()
	consumerConfig.ClientID = ks.ClientID // should be config
	consumerConfig.Version = sarama.V1_0_0_0
	consumerConfig.Net.TLS.Enable = true
	consumerConfig.Net.TLS.Config = tlsConfig
	consumerConfig.Consumer.Return.Errors = true

	// Start with a client
	client, err := sarama.NewClient(ks.Brokers, consumerConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	// Start a new consumer group
	group, err := sarama.NewConsumerGroupFromClient(ks.GroupName, client)
	if err != nil {
		log.Print(err)
	}
	defer func() { _ = group.Close() }()

	// go context, not our own
	context := context.Background()

	// TODO: how to make channel sending function return an error?
	h := func(um UpdateMessage) {
		updates <- um
	}

	//for {
	handler := consumerGroupHandler{
		HandleUpdateMessage: h,
	}
	// NOTE: was in loop
	//err := group.Consume(ctx, []string{ks.ChangesTopic}, handler)
	err = group.Consume(context, ks.Topics, handler)
	if err != nil {
		log.Print(err)
	}
	//}
	// how to return channel?
	return updates
}
*/

/*
func (ks KafkaSubscriber) SubscribeOld(ctx Context) chan UpdateMessage {
	updates := make(chan UpdateMessage)

	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	brokers := ks.Brokers
	topics := ks.Topics

	ctx.Logger.Printf("%s:%s:%s", ks.ClientCert, ks.ClientKey, ks.ServerCert)

	tlsConfig, err := NewTLSConfig(ks.ClientCert, ks.ClientKey, ks.ServerCert)

	if err != nil {
		log.Fatal(err)
	}

	consumerConfig := sarama.NewConfig()
	consumerConfig.ClientID = ks.ClientID // should be config
	consumerConfig.Net.TLS.Enable = true
	consumerConfig.Net.TLS.Config = tlsConfig
	consumerConfig.Consumer.Return.Errors = true

	// Start with a client
	client, err := sarama.NewClient(ks.Brokers, consumerConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	// Start a new consumer group
	group, err := sarama.NewConsumerGroupFromClient(ks.GroupName, client)
	if err != nil {
		log.Print(err)
	}
	defer func() { _ = group.Close() }()




		consumer, err := cluster.NewConsumer(
			brokers,
			"group-id",
			topics,
			consumerConfig)

		if err != nil {
			panic(err)
			//close(updates)
			//ctx.handleError("Problem communicating with kafka queue", err, true)
		}

		defer consumer.Close()



	go func() {
		// The loop will iterate each time a message is written to the underlying channel
		for msg, ok := range group.Messages() {
			//for msg, ok := range consumer.Messages() {
			var m UpdateMessage
			json.Unmarshal(msg.Value, &m)
			updates <- m
			log.Printf("REC:uri=%s", m.Triple.Subject)
			// TODO: way to apply backpressure here - or somewhere?
			// Mark the message as processed. The sarama-cluster library will
			// automatically commit these.
			// You can manually commit the offsets using consumer.CommitOffsets()
			consumer.MarkOffset(msg, "required-metadata")
			//break
		}
	}()
	return updates
}
*/
