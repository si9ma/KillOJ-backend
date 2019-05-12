package srv

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"

	"github.com/jinzhu/copier"

	"github.com/si9ma/KillOJ-backend/wrap"

	"github.com/si9ma/KillOJ-common/kjson"

	uuid "github.com/satori/go.uuid"

	"github.com/si9ma/KillOJ-backend/data"

	"github.com/si9ma/KillOJ-common/kredis"

	"github.com/si9ma/KillOJ-common/utils"

	"github.com/si9ma/KillOJ-backend/kerror"

	"github.com/si9ma/KillOJ-backend/auth"

	"github.com/jinzhu/gorm"

	"go.uber.org/zap"

	"github.com/si9ma/KillOJ-common/log"
	"github.com/si9ma/KillOJ-common/mysql"

	"github.com/si9ma/KillOJ-backend/gbl"

	"github.com/gin-gonic/gin"

	"github.com/si9ma/KillOJ-common/model"
	otgrom "github.com/smacker/opentracing-gorm"
)

const (
	ContestInvitePrefix = "contest_invite_"
)

func GetAllContests(c *gin.Context, page, pageSize int, order string) ([]model.Contest, error) {
	var err error

	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	offset := (page - 1) * pageSize
	user := model.User{
		ID: auth.GetUserFromJWT(c).ID,
	}

	if order != "" {
		err = db.Preload("Contests", func(db *gorm.DB) *gorm.DB {
			return db.Order(order).Offset(offset).Limit(pageSize)
		}).First(&user).Error
	} else {
		err = db.Preload("Contests", func(db *gorm.DB) *gorm.DB {
			return db.Offset(offset).Limit(pageSize)
		}).First(&user).Error
	}

	// error handle
	if mysql.ErrorHandleAndLog(c, err, true,
		"get contests", nil) != mysql.Success {
		return nil, err
	}

	log.For(ctx).Info("success get contests")
	return user.Contests, nil
}

func GetContest(c *gin.Context, id int) (*model.Contest, error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	user := model.User{
		ID: auth.GetUserFromJWT(c).ID,
	}

	// get contest
	err := db.Preload("Contests", "contest_id = ?", id).First(&user).Error
	if mysql.ErrorHandleAndLog(c, err, true, "get contest", id) != mysql.Success {
		return nil, err
	}
	// check contest exist
	if len(user.Contests) == 0 {
		log.For(ctx).Error("no contest or user not in contest",
			zap.Int("contestId", id), zap.Int("userId", user.ID))

		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrNotFound.WithArgs(id))
		return nil, kerror.EmptyError
	}
	contest := user.Contests[0]

	log.For(ctx).Info("success get contest", zap.Int("contestId", id))
	return &contest, nil
}

func AddContest(c *gin.Context, newContest *model.Contest) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	oldContest := model.Contest{}

	//  check unique
	if !contestCheckUnique(c, &oldContest, newContest) {
		return fmt.Errorf("check unique fail")
	}

	// add new contest
	user := model.User{ID: newContest.OwnerID}
	err := db.Model(&user).Association("Contests").Append(newContest).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"add new contest", newContest.Name) != mysql.Success {
		return err
	}

	log.For(ctx).Info("add new contest success",
		zap.String("contestName", newContest.Name))
	return nil
}

func UpdateContest(c *gin.Context, newContest *model.Contest) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// check if contest exist
	oldContest, err := GetContest(c, newContest.ID)
	if err != nil {
		return err
	}

	// check owner
	if err := checkContestOwner(c, oldContest); err != nil {
		return err
	}

	//  check unique
	if !contestCheckUnique(c, oldContest, newContest) {
		return fmt.Errorf("check unique fail")
	}

	// update
	err = db.Model(oldContest).Updates(newContest).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"update contest", newContest.ID) != mysql.Success {
		return err
	}
	log.For(ctx).Info("update contest success", zap.String("contest", newContest.Name))

	return nil
}

func contestCheckUnique(c *gin.Context, oldContest *model.Contest, newContest *model.Contest) bool {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// unique check
	// new value -- old value
	uniqueCheckList := []map[string]mysql.ValuePair{
		{
			"name": mysql.ValuePair{
				NewVal: newContest.Name,
				OldVal: oldContest.Name,
			},
		},
	}
	for _, checkMap := range uniqueCheckList {
		if !mysql.ShouldUnique(c, ctx, db, checkMap, func(db *gorm.DB) error {
			return db.First(&model.Contest{}).Error
		}) {
			return false
		}
	}

	return true
}

func checkContestOwner(c *gin.Context, contest *model.Contest) error {
	ctx := c.Request.Context()

	// check owner
	userId := auth.GetUserFromJWT(c).ID
	if userId != contest.OwnerID {
		err := fmt.Errorf("operate contest forbidden")
		log.For(ctx).Error("operate contest fail(forbidden)", zap.Int("contestId", contest.ID))

		_ = c.Error(err).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrForbiddenGeneral)
		return err
	}

	return nil
}

//func DeleteContest(c *gin.Context, id int) error {
//	ctx := c.Request.Context()
//	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
//
//	// check if contest exist
//	contest, err := GetContest(c, id)
//	if err != nil {
//		return err
//	}
//
//	// check owner
//	if err := checkGroupOwner(c, contest); err != nil {
//		return err
//	}
//	// todo should check have problem under this contest
//
//	err = db.Delete(&contest).Error
//	if mysql.ErrorHandleAndLog(c, err, true,
//		"delete contest", id) != mysql.Success {
//		return err
//	}
//	log.For(ctx).Info("delete contest success", zap.Int("contestId", id))
//
//	return nil
//}

// check if user have these group Permission
func CheckPermission(c *gin.Context, groups []int, anyOne bool) error {
	var err error

	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	user := model.User{
		ID: auth.GetUserFromJWT(c).ID,
	}

	err = db.Preload("Groups", "group_id in (?)", groups).First(&user).Error

	// error handle
	if mysql.ErrorHandleAndLog(c, err, true,
		"get groups", nil) != mysql.Success {
		return err
	}

	// if anyone match
	if anyOne && len(user.Groups) > 0 {
		// ok
		log.For(ctx).Info("check permission success")
		return nil
	}

	// all

	// user not all these Permission
	if len(user.Groups) != len(groups) {
		var failGroups []int
		var okGroups []int

		for _, group := range user.Groups {
			okGroups = append(okGroups, group.ID)
		}
		for _, group := range groups {
			// if not contains
			if !utils.ContainsInt(okGroups, group) {
				failGroups = append(failGroups, group)
			}
		}
		fields := map[string]interface{}{
			"groups": failGroups,
		}
		log.For(ctx).Error("user not these permission", zap.Any("groups", failGroups))

		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrForbiddenGeneral.With(fields))
		return fmt.Errorf("check groups permission fail")
	}

	log.For(ctx).Info("check all permission success", zap.Any("groups", groups))
	return nil
}

func GetContestInviteInfo(c *gin.Context, contestID int) (*data.ContestInviteData, error) {
	ctx := c.Request.Context()
	redisCli := kredis.WrapRedisClient(ctx, gbl.Redis)

	// check if contest exist
	contest, err := GetContest(c, contestID)
	if err != nil {
		return nil, err
	}

	// check owner
	if err := checkContestOwner(c, contest); err != nil {
		return nil, err
	}

	timeout := contest.EndTime.Sub(time.Now())
	if timeout <= 0 {
		// contest already finish
		log.For(ctx).Error("contest already finished", zap.Int("contestID", contest.ID))
		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrAlreadyFinished)
		return nil, fmt.Errorf("contest already finished")
	}

	// check if already invite,
	// if already invite --> return data
	k := ContestInvitePrefix + strconv.Itoa(contestID)
	uuidVal, err := redisCli.Get(k).Result()
	res := kredis.ErrorHandleAndLog(c, err, false,
		"check if already invite contest", k, contestID)
	switch res {
	case kredis.Success:
		// remove prefix
		inviteId := strings.ReplaceAll(uuidVal, ContestInvitePrefix, "")
		d, err := GetContestInviteData(c, inviteId)
		return d, err
	default:
		return nil, err
	}
}

func Invite2Contest(c *gin.Context, inviteData *data.ContestInviteData) (err error) {
	ctx := c.Request.Context()
	redisCli := kredis.WrapRedisClient(ctx, gbl.Redis)

	d, err := GetContestInviteInfo(c, inviteData.ContestID)
	if err != nil && err == redis.Nil {
		// because we store invite data in redis --> if error is redis.Nil,
		// invite not already exist, we should generate new invite info,
		// so continue
	} else {
		// if other error or invite already exist, we should copy invite date to inviteData
		// then return
		copier.Copy(inviteData, d)
		return err
	}

	// get contest info
	contest, err := GetContest(c, inviteData.ContestID)
	if err != nil {
		return err
	}

	// check permission
	if len(inviteData.AllowGroups) != 0 {
		if err := CheckPermission(c, inviteData.AllowGroups, false); err != nil {
			return err
		}
	}

	// generate uuid
	uuid, err := uuid.NewV4()
	if err != nil {
		log.For(ctx).Error("generate uuid fail", zap.Error(err))

		wrap.SetInternalServerError(c, err)
		return
	}
	inviteData.ID = uuid.String()

	// save invite data to redis
	res, err := kjson.MarshalString(inviteData)
	if err != nil {
		log.For(ctx).Error("marshal json fail", zap.Error(err),
			zap.Int("contestID", inviteData.ContestID))

		wrap.SetInternalServerError(c, err)
		return err
	}

	k1, k2 := ContestInvitePrefix+inviteData.ID, ContestInvitePrefix+strconv.Itoa(inviteData.ContestID)

	// save invite data
	timeout := contest.EndTime.Sub(time.Now())
	err = redisCli.Set(k1, res, timeout).Err()
	if res := kredis.ErrorHandleAndLog(c, err, true,
		"save contest invite data", k1, nil); res != kredis.Success {
		return err
	}

	// save k1 key as the value of k2
	err = redisCli.Set(k2, k1, timeout).Err()
	if res := kredis.ErrorHandleAndLog(c, err, true,
		"save contest invite data", k1, nil); res != kredis.Success {
		return err
	}

	return nil
}

func GetContestInviteData(c *gin.Context, inviteId string) (*data.ContestInviteData, error) {
	ctx := c.Request.Context()
	redisCli := kredis.WrapRedisClient(ctx, gbl.Redis)

	// check if inviteId available
	k := ContestInvitePrefix + inviteId
	res, err := redisCli.Get(k).Result()
	if r := kredis.ErrorHandleAndLog(c, err, false,
		"get invite data", k, nil); r == kredis.NotFound {
		return nil, err
	} else if r != kredis.Success {
		return nil, err
	}

	inviteData := data.ContestInviteData{}
	if err := kjson.UnmarshalString(res, &inviteData); err != nil {
		log.For(ctx).Error("unmarshal fail", zap.Error(err))
		wrap.SetInternalServerError(c, err)
		return nil, err
	}

	return &inviteData, nil
}

// query before join
func JoinContestQuery(c *gin.Context, inviteId string) (*data.ContestWrap, error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	inviteData, err := GetContestInviteData(c, inviteId)
	if err != nil {
		return nil, err
	}

	contest := model.Contest{}
	err = db.First(&contest, inviteData.ContestID).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"get contest", inviteData.ContestID) != mysql.Success {
		return nil, err
	}

	needPassword := false
	// if check permission success, not need password
	if err := CheckPermission(c, inviteData.AllowGroups, true); err != nil {
		needPassword = inviteData.Password != "" // if password is not empty, need password to join
	}
	return &data.ContestWrap{
		Contest:      contest,
		NeedPassword: needPassword,
		Password:     inviteData.Password,
	}, nil
}

func JoinContest(c *gin.Context, inviteId string, password string) (err error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// check if contest exist
	contestWrap, err := JoinContestQuery(c, inviteId)
	if err != nil {
		return err
	}

	// if need password
	if contestWrap.NeedPassword && password != contestWrap.Password {
		log.For(ctx).Error("join contest password wrong", zap.Int("contestID", contestWrap.ID))
		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrPasswordWrong)
		return fmt.Errorf("join contest password wrong")
	}

	contest := contestWrap.Contest
	user := model.User{
		ID: auth.GetUserFromJWT(c).ID,
	}
	err = db.Model(&user).Association("Contests").Append(&contest).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"add user to contest", contest.ID) != mysql.Success {
		return err
	}

	return nil
}
