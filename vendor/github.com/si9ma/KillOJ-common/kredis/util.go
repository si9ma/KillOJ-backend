package kredis

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/si9ma/KillOJ-backend/kerror"
	"github.com/si9ma/KillOJ-common/log"
	"github.com/si9ma/KillOJ-common/model"
	"github.com/si9ma/KillOJ-common/utils"
	"go.uber.org/zap"
)

func GetTestRedis() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6666",
	})

	_, err := client.Ping().Result()
	return client, err
}

type ErrHandleResult int

const (
	Success = ErrHandleResult(iota)
	NotFound
	DB_ERROR
)

// desc and key for log
func ErrorHandleAndLog(c *gin.Context, err error, treatNotFoundAsErr bool, desc string, key string, extra interface{}) ErrHandleResult {
	ctx := c.Request.Context()

	// get myself
	myself, ok := utils.SafeGetUserFromJWT(c)
	if !ok {
		myself = model.User{}
	}

	if err == redis.Nil {
		if treatNotFoundAsErr {
			// set gin error
			log.For(ctx).Error("key not exist", zap.Error(err), zap.Any("key", key),
				zap.String("desc", desc), zap.Any("extra", extra), zap.Int("userID", myself.ID))

			_ = c.Error(err).SetType(gin.ErrorTypePublic).
				SetMeta(kerror.ErrNotFoundOrOutOfDate.WithArgs(extra))
		}
		// not set gin error
		return NotFound
	} else if err != nil {
		log.For(ctx).Error("operate redis fail", zap.Error(err), zap.Any("key", key),
			zap.String("desc", desc), zap.Any("extra", extra), zap.Int("userID", myself.ID))

		_ = c.Error(err).SetType(gin.ErrorTypePublic).SetMeta(kerror.ErrInternalServerErrorGeneral)
		return DB_ERROR
	}

	return Success
}
