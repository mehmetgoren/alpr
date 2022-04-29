package reps

import (
	"alpr/utils"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
	"log"
	"time"
)

type HeartbeatRepository struct {
	Client      *redis.Client
	TimeSecond  int64
	ServiceName string
}

func (h *HeartbeatRepository) Start() {
	var dur = time.Duration(h.TimeSecond) * time.Second
	ticker := time.NewTicker(dur)
	for {
		select {
		case timeTicker := <-ticker.C:
			heartbeatObj := map[string]interface{}{
				"heartbeat": utils.TimeToString(timeTicker, true),
			}
			h.Client.HSet(context.Background(), "services:"+h.ServiceName, heartbeatObj)
			log.Println("Heartbeat was beaten at " + timeTicker.Format(time.ANSIC))
		}
	}
}
