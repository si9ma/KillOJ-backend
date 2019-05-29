package mysql

import (
	"context"

	"github.com/si9ma/KillOJ-common/model"

	"github.com/si9ma/KillOJ-common/utils"

	"github.com/si9ma/KillOJ-backend/kerror"
	"github.com/si9ma/KillOJ-common/log"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func GetTestDB() (db *gorm.DB, err error) {
	cfg := Config{
		ConnectionStr: "root:mysqlpass@tcp(127.0.0.1:3306)/killoj?&parseTime=True",
	}
	return InitDB(cfg)
}

type ErrHandleResult int

const (
	Success = ErrHandleResult(iota)
	NotFound
	DB_ERROR
)

// desc and extra for log
func ErrorHandleAndLog(c *gin.Context, err error, treatNotFoundAsErr bool, desc string, extra ...interface{}) ErrHandleResult {
	ctx := c.Request.Context()

	// get myself
	myself, ok := utils.SafeGetUserFromJWT(c)
	if !ok {
		myself = model.User{}
	}

	if gorm.IsRecordNotFoundError(err) {
		if treatNotFoundAsErr {
			// set gin error
			log.For(ctx).Error("record not exist", zap.Error(err), zap.Any("extra", extra),
				zap.String("desc", desc), zap.Int("userID", myself.ID))

			_ = c.Error(err).SetType(gin.ErrorTypePublic).
				SetMeta(kerror.ErrNotFound.WithArgs(extra))
		}
		// not set gin error
		return NotFound
	} else if err != nil {
		log.For(ctx).Error("operate db fail", zap.Error(err), zap.Any("extra", extra),
			zap.String("desc", desc), zap.Int("userID", myself.ID))

		_ = c.Error(err).SetType(gin.ErrorTypePublic).SetMeta(kerror.ErrInternalServerErrorGeneral)
		return DB_ERROR
	}

	return Success
}

type ValuePair struct {
	NewVal interface{}
	OldVal interface{}
}

// fields should unique
func ShouldUnique(c *gin.Context, ctx context.Context, db *gorm.DB, fieldMap map[string]ValuePair, operate func(*gorm.DB) error) bool {
	isSame := true
	checkMap := make(map[string]interface{})

	for k, v := range fieldMap {
		// is some field is zero, don't check unique
		if utils.IsZeroOfUnderlyingType(v.NewVal) {
			log.For(ctx).Info("some field is zero",
				zap.String("field", k), zap.Any("value", v))
			return true
		}

		// value changed
		if v.NewVal != v.OldVal {
			isSame = false
		}
		checkMap[k] = v.NewVal
	}

	// if value is same,
	// don't check unique
	if isSame {
		log.For(ctx).Info("same values, don't check unique", zap.Any("fieldMap", fieldMap))
		return true
	}

	// because gorm not allow use interface as destination,
	// so , use callback function to do real db operate,
	// err := db.Where(queryStr, fieldVals...).First(&object).Error
	db = db.Where(checkMap)
	err := operate(db)
	if res := ErrorHandleAndLog(c, err, false,
		"user unique check", checkMap); res == Success {
		// already exist
		log.For(ctx).Error("fields already exist", zap.Any("fields", checkMap))

		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrAlreadyExist.WithArgs(checkMap).With(checkMap))
		return false
	} else if res == NotFound {
		return true
	}

	return false
}
