package vivoupdater

/*
import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"time"
)

type UpdateSubscriber struct {
	RedisUrl           string
	RedisChannel       string
	MaxConnectAttempts int
	RetryInterval      int
}

func (us UpdateSubscriber) Subscribe(ctx Context) chan UpdateMessage {
	updates := make(chan UpdateMessage)
	go func() {
		failedAttempts := 0
		for {
			ctx.Logger.Println("Connecting to Redis at " + us.RedisUrl)
			c, err := redis.Dial("tcp", us.RedisUrl)
			if err != nil {
				failedAttempts += 1
				if failedAttempts > us.MaxConnectAttempts {
					close(updates)
					ctx.handleError("Max redis connection attempts exceeded", err, true)
					break
				}
				ctx.handleError("Could not connect to Redis", err, false)
				time.Sleep(time.Duration(us.RetryInterval*failedAttempts) * time.Second)
				continue
			}
			failedAttempts = 0
			psc := redis.PubSubConn{c}
			psc.Subscribe(us.RedisChannel)
			ctx.Logger.Println("Subscribed to channel " + us.RedisChannel)

		ReceiveLoop:
			for {
				switch v := psc.Receive().(type) {
				case redis.Message:
					var m UpdateMessage
					json.Unmarshal(v.Data, &m)
					updates <- m
				case error:
					ctx.handleError("Lost connection to Redis", v, false)
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
*/
