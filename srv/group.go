package srv

import (
	"fmt"
	"strconv"
	"time"

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
	InvitePrefix = "killoj_invite_"
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
	if err := checkOwner(c, oldGroup); err != nil {
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

func checkOwner(c *gin.Context, group *model.Group) error {
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
//	if err := checkOwner(c, group); err != nil {
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

// check if user have these group Permission
func CheckPermission(c *gin.Context, groups []int) error {
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

func Invite(c *gin.Context, inviteData *data.InviteData) (err error) {
	ctx := c.Request.Context()
	redisCli := kredis.WrapRedisClient(ctx, gbl.Redis)

	// check if group exist
	oldGroup, err := GetGroup(c, inviteData.GroupID)
	if err != nil {
		return err
	}

	// check owner
	if err := checkOwner(c, oldGroup); err != nil {
		return err
	}
	//
	//// check permission
	//if len(inviteData.AllowGroups) != 0 {
	//	if err := CheckPermission(c, inviteData.AllowGroups); err != nil {
	//		return err
	//	}
	//}

	// check if already invite
	k := InvitePrefix + strconv.Itoa(inviteData.GroupID)
	err = redisCli.Get(k).Err()
	if res := kredis.ErrorHandleAndLog(c, err, false,
		"check if already invite group", k, inviteData.GroupID); res == kredis.Success {

		err := fmt.Errorf("already invite group")
		_ = c.Error(err).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrAlreadyInvite)
		return err
	} else if res != kredis.NotFound {
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

	// save invite data to redisCli
	res, err := kjson.MarshalString(inviteData)
	if err != nil {
		log.For(ctx).Error("marshal json fail", zap.Error(err),
			zap.Int("groupID", inviteData.GroupID))

		wrap.SetInternalServerError(c, err)
		return err
	}

	timeout := time.Duration(inviteData.Timeout) * time.Second
	k1, k2 := InvitePrefix+inviteData.ID, InvitePrefix+strconv.Itoa(inviteData.GroupID)

	// save invite data
	err = redisCli.Set(k1, res, timeout).Err()
	if res := kredis.ErrorHandleAndLog(c, err, true,
		"save invite data", k1, nil); res != kredis.Success {
		return err
	}

	// mark group as already invite
	err = redisCli.Set(k2, true, timeout).Err()
	if res := kredis.ErrorHandleAndLog(c, err, true,
		"save invite data", k1, nil); res != kredis.Success {
		return err
	}

	return nil
}

// query before join
func JoinQuery(c *gin.Context, inviteId string) (*data.GroupWrap, error) {
	ctx := c.Request.Context()
	redisCli := kredis.WrapRedisClient(ctx, gbl.Redis)
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// check if inviteId available
	k := InvitePrefix + inviteId
	res, err := redisCli.Get(k).Result()
	if kredis.ErrorHandleAndLog(c, err, false,
		"get invite data", k, nil) != kredis.Success {
		return nil, err
	}

	inviteData := data.InviteData{}
	if err := kjson.UnmarshalString(res, &inviteData); err != nil {
		log.For(ctx).Error("unmarshal fail", zap.Error(err))
		wrap.SetInternalServerError(c, err)
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

func Join(c *gin.Context, inviteId string, password string) (err error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// check if group exist
	groupWrap, err := JoinQuery(c, inviteId)
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
