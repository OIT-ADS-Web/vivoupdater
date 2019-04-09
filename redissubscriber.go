package vivoupdater

// NOTE: not using redis anymore
import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

type RedisSubscriber struct {
	RedisUrl           string
	RedisChannel       string
	MaxConnectAttempts int
	RetryInterval      int
}

func (us RedisSubscriber) Subscribe(ctx context.Context, logger *log.Logger) chan UpdateMessage {
	updates := make(chan UpdateMessage)
	_, cancel := context.WithCancel(ctx)
	go func() {
		failedAttempts := 0
		for {
			logger.Println("Connecting to Redis at " + us.RedisUrl)
			c, err := redis.Dial("tcp", us.RedisUrl)
			if err != nil {
				failedAttempts += 1
				if failedAttempts > us.MaxConnectAttempts {
					close(updates)
					logger.Printf("Max redis connection attempts exceeded: %v\n", err)
					cancel()
					break
				}
				logger.Printf("Could not connect to Redis: %v\n", err)
				time.Sleep(time.Duration(us.RetryInterval*failedAttempts) * time.Second)
				continue
			}
			failedAttempts = 0
			psc := redis.PubSubConn{c}
			psc.Subscribe(us.RedisChannel)
			logger.Println("Subscribed to channel " + us.RedisChannel)

		ReceiveLoop:
			for {
				switch v := psc.Receive().(type) {
				case redis.Message:
					var m UpdateMessage
					json.Unmarshal(v.Data, &m)
					updates <- m
				case error:
					logger.Printf("Lost connection to Redis: %v\n", v)
					psc.Close()
					time.Sleep(time.Duration(us.RetryInterval) * time.Second)
					failedAttempts += 1
					break ReceiveLoop
				}
			}
		}
	}()
	return updates
}
