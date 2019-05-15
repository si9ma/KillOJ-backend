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
	GroupInvitePrefix = "group_invite_"
)

func GetAllGroups(c *gin.Context, page, pageSize int, order string) ([]model.Group, error) {
	var err error

	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	offset := (page - 1) * pageSize
	user := model.User{
		ID: auth.GetUserFromJWT(c).ID,
	}

	if order != "" {
		err = db.Preload("Groups", func(db *gorm.DB) *gorm.DB {
			return db.Order(order).Offset(offset).Limit(pageSize)
		}).First(&user).Error
	} else {
		err = db.Preload("Groups", func(db *gorm.DB) *gorm.DB {
			return db.Offset(offset).Limit(pageSize)
		}).First(&user).Error
	}

	// error handle
	if mysql.ErrorHandleAndLog(c, err, true,
		"get groups", nil) != mysql.Success {
		return nil, err
	}

	log.For(ctx).Info("success get groups")
	return user.Groups, nil
}

func GetGroup(c *gin.Context, id int) (*model.Group, error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	user := model.User{
		ID: auth.GetUserFromJWT(c).ID,
	}

	// get group
	err := db.Preload("Groups", "group_id = ?", id).First(&user).Error
	if mysql.ErrorHandleAndLog(c, err, true, "get group", id) != mysql.Success {
		return nil, err
	}
	// check group exist
	if len(user.Groups) == 0 {
		log.For(ctx).Error("no group or user not in group",
			zap.Int("groupId", id), zap.Int("userId", user.ID))

		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrNotFound.WithArgs(id))
		return nil, kerror.EmptyError
	}
	group := user.Groups[0]

	log.For(ctx).Info("success get group", zap.Int("groupId", id))
	return &group, nil
}

func AddGroup(c *gin.Context, newGroup *model.Group) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	oldGroup := model.Group{}

	//  check unique
	if !groupCheckUnique(c, &oldGroup, newGroup) {
		return fmt.Errorf("check unique fail")
	}

	// add new group
	user := model.User{ID: newGroup.OwnerID}
	err := db.Model(&user).Association("Groups").Append(newGroup).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"add new group", newGroup.Name) != mysql.Success {
		return err
	}

	log.For(ctx).Info("add new group success",
		zap.String("groupName", newGroup.Name))
	return nil
}

func UpdateGroup(c *gin.Context, newGroup *model.Group) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// check if group exist
	oldGroup, err := GetGroup(c, newGroup.ID)
	if err != nil {
		return err
	}

	// check owner
	if err := checkGroupOwner(c, oldGroup); err != nil {
		return err
	}

	//  check unique
	if !groupCheckUnique(c, oldGroup, newGroup) {
		return fmt.Errorf("check unique fail")
	}

	// update
	err = db.Model(oldGroup).Updates(newGroup).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"update group", newGroup.ID) != mysql.Success {
		return err
	}
	log.For(ctx).Info("update group success", zap.String("group", newGroup.Name))

	return nil
}

func groupCheckUnique(c *gin.Context, oldGroup *model.Group, newGroup *model.Group) bool {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// unique check
	// new value -- old value
	uniqueCheckList := []map[string]mysql.ValuePair{
		{
			"name": mysql.ValuePair{
				NewVal: newGroup.Name,
				OldVal: oldGroup.Name,
			},
		},
	}
	for _, checkMap := range uniqueCheckList {
		if !mysql.ShouldUnique(c, ctx, db, checkMap, func(db *gorm.DB) error {
			return db.First(&model.Group{}).Error
		}) {
			return false
		}
	}

	return true
}

func checkGroupOwner(c *gin.Context, group *model.Group) error {
	ctx := c.Request.Context()

	// check owner
	userId := auth.GetUserFromJWT(c).ID
	if userId != group.OwnerID {
		err := fmt.Errorf("operate group forbidden")
		log.For(ctx).Error("operate group fail(forbidden)", zap.Int("groupId", group.ID))

		_ = c.Error(err).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrForbiddenGeneral)
		return err
	}

	return nil
}

//func DeleteGroup(c *gin.Context, id int) error {
//	ctx := c.Request.Context()
//	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
//
//	// check if group exist
//	group, err := GetGroup(c, id)
//	if err != nil {
//		return err
//	}
//
//	// check owner
//	if err := checkGroupOwner(c, group); err != nil {
//		return err
//	}
//	// todo should check have problem under this group
//
//	err = db.Delete(&group).Error
//	if mysql.ErrorHandleAndLog(c, err, true,
//		"delete group", id) != mysql.Success {
//		return err
//	}
//	log.For(ctx).Info("delete group success", zap.Int("groupId", id))
//
//	return nil
//}

func GetGroupInviteInfo(c *gin.Context, groupID int) (*data.GroupInviteData, error) {
	ctx := c.Request.Context()
	redisCli := kredis.WrapRedisClusterClient(ctx, gbl.Redis)

	// check if group exist
	group, err := GetGroup(c, groupID)
	if err != nil {
		return nil, err
	}

	// check owner
	if err := checkGroupOwner(c, group); err != nil {
		return nil, err
	}

	// check if already invite,
	// if already invite --> return data
	k := GroupInvitePrefix + strconv.Itoa(groupID)
	uuidVal, err := redisCli.Get(k).Result()
	res := kredis.ErrorHandleAndLog(c, err, false,
		"check if already invite group", k, groupID)
	switch res {
	case kredis.Success:
		// remove prefix
		inviteId := strings.ReplaceAll(uuidVal, GroupInvitePrefix, "")
		d, err := GetGroupInviteData(c, inviteId)
		return d, err
	default:
		return nil, err
	}
}

func Invite2Group(c *gin.Context, inviteData *data.GroupInviteData) (err error) {
	ctx := c.Request.Context()
	redisCli := kredis.WrapRedisClusterClient(ctx, gbl.Redis)

	d, err := GetGroupInviteInfo(c, inviteData.GroupID)
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
			zap.Int("groupID", inviteData.GroupID))

		wrap.SetInternalServerError(c, err)
		return err
	}

	timeout := time.Duration(inviteData.Timeout) * time.Second
	k1, k2 := GroupInvitePrefix+inviteData.ID, GroupInvitePrefix+strconv.Itoa(inviteData.GroupID)

	// save invite data
	err = redisCli.Set(k1, res, timeout).Err()
	if res := kredis.ErrorHandleAndLog(c, err, true,
		"save group invite data", k1, nil); res != kredis.Success {
		return err
	}

	// mark group as already invite
	err = redisCli.Set(k2, true, timeout).Err()
	if res := kredis.ErrorHandleAndLog(c, err, true,
		"save group invite data", k1, nil); res != kredis.Success {
		return err
	}

	return nil
}

// return value:
func GetGroupInviteData(c *gin.Context, inviteId string) (*data.GroupInviteData, error) {
	ctx := c.Request.Context()
	redisCli := kredis.WrapRedisClusterClient(ctx, gbl.Redis)

	// check if inviteId available
	k := GroupInvitePrefix + inviteId
	res, err := redisCli.Get(k).Result()
	if r := kredis.ErrorHandleAndLog(c, err, false,
		"get invite data", k, nil); r != kredis.NotFound {
		return nil, err
	}

	inviteData := data.GroupInviteData{}
	if err := kjson.UnmarshalString(res, &inviteData); err != nil {
		log.For(ctx).Error("unmarshal fail", zap.Error(err))
		wrap.SetInternalServerError(c, err)
		return nil, err
	}

	return &inviteData, nil
}

// query before join
func JoinGroupQuery(c *gin.Context, inviteId string) (*data.GroupWrap, error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	inviteData, err := GetGroupInviteData(c, inviteId)
	if err != nil {
		return nil, err
	}

	group := model.Group{}
	err = db.First(&group, inviteData.GroupID).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"get group", inviteData.GroupID) != mysql.Success {
		return nil, err
	}

	return &data.GroupWrap{
		Group:        group,
		NeedPassword: inviteData.Password != "", // if password is not empty, need password to join
		Password:     inviteData.Password,
	}, nil
}

func JoinGroup(c *gin.Context, inviteId string, password string) (err error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// check if group exist
	groupWrap, err := JoinGroupQuery(c, inviteId)
	if err != nil {
		return err
	}

	// if need password
	if groupWrap.NeedPassword && password != groupWrap.Password {
		log.For(ctx).Error("join group password wrong", zap.Int("groupID", groupWrap.ID))
		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrPasswordWrong)
		return fmt.Errorf("join group password wrong")
	}

	group := groupWrap.Group
	user := model.User{
		ID: auth.GetUserFromJWT(c).ID,
	}
	err = db.Model(&user).Association("Groups").Append(&group).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"add user to group", group.ID) != mysql.Success {
		return err
	}

	return nil
}
