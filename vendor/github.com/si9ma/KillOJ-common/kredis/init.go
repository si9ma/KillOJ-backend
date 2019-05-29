// redis wrapper
package kredis

import (
	"github.com/go-redis/redis"
	"github.com/si9ma/KillOJ-common/utils"
)

func Init(cfg Config) (*redis.ClusterClient, error) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        cfg.Addrs,
		DialTimeout:  utils.Millisecond(cfg.DialTimeout),
		ReadTimeout:  utils.Millisecond(cfg.ReadTimeout),
		WriteTimeout: utils.Millisecond(cfg.WriteTimeout),
	})

	_, err := client.Ping().Result()
	return client, err
}
