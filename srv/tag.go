package srv

import (
	"fmt"

	"github.com/jinzhu/gorm"

	"go.uber.org/zap"

	"github.com/si9ma/KillOJ-common/log"
	"github.com/si9ma/KillOJ-common/mysql"

	"github.com/si9ma/KillOJ-backend/gbl"

	"github.com/gin-gonic/gin"

	"github.com/si9ma/KillOJ-common/model"
	otgrom "github.com/smacker/opentracing-gorm"
)

func GetAllTags(c *gin.Context) ([]model.Tag, error) {
	var err error
	var tags []model.Tag

	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// error handle
	err = db.Find(&tags).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"get tags", nil) != mysql.Success {
		return nil, err
	}

	log.For(ctx).Info("success get tags")
	return tags, nil
}

func GetTag(c *gin.Context, id int) (*model.Tag, error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	tag := &model.Tag{}
	err := db.First(&tag, id).Error
	if mysql.ErrorHandleAndLog(c, err, true, "get tag", id) != mysql.Success {
		return nil, err
	}

	log.For(ctx).Info("success get tag", zap.Int("tagId", id))
	return tag, nil
}

func AddTag(c *gin.Context, newTag *model.Tag) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	oldTag := model.Tag{}

	//  check unique
	if !tagCheckUnique(c, &oldTag, newTag) {
		return fmt.Errorf("check unique fail")
	}

	err := db.Create(&newTag).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"add new tag", newTag.Name) != mysql.Success {
		return err
	}

	log.For(ctx).Info("add new tag success",
		zap.String("tagName", newTag.Name))
	return nil
}

func UpdateTag(c *gin.Context, newTag *model.Tag) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// check if tag exist
	oldTag, err := GetTag(c, newTag.ID)
	if err != nil {
		return err
	}

	//  check unique
	if !tagCheckUnique(c, oldTag, newTag) {
		return fmt.Errorf("check unique fail")
	}

	// update
	err = db.Model(oldTag).Updates(newTag).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"update tag", newTag.ID) != mysql.Success {
		return err
	}
	log.For(ctx).Info("update tag success", zap.String("tag", newTag.Name))

	return nil
}

func tagCheckUnique(c *gin.Context, oldTag *model.Tag, newTag *model.Tag) bool {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// unique check
	// new value -- old value
	uniqueCheckList := []map[string]mysql.ValuePair{
		{
			"name": mysql.ValuePair{
				NewVal: newTag.Name,
				OldVal: oldTag.Name,
			},
		},
	}
	for _, checkMap := range uniqueCheckList {
		if !mysql.ShouldUnique(c, ctx, db, checkMap, func(db *gorm.DB) error {
			return db.First(&model.Tag{}).Error
		}) {
			return false
		}
	}

	return true
}

func DeleteTag(c *gin.Context, id int) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// check if tag exist
	tag, err := GetTag(c, id)
	if err != nil {
		return err
	}

	// todo should check have problem under this tag

	err = db.Delete(&tag).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"delete tag", id) != mysql.Success {
		return err
	}
	log.For(ctx).Info("delete tag success", zap.Int("tagId", id))

	return nil
}
