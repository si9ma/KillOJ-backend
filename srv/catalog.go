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

func GetAllCatalogs(c *gin.Context, page, pageSize int, order string) ([]model.Catalog, error) {
	var err error

	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	offset := (page - 1) * pageSize

	var catalogs []model.Catalog

	if order != "" {
		err = db.Model(&model.Catalog{}).Order(order).Offset(offset).Limit(pageSize).Find(&catalogs).Error
	} else {
		err = db.Model(&model.Catalog{}).Offset(offset).Limit(pageSize).Find(&catalogs).Error
	}

	// error handle
	if mysql.ErrorHandleAndLog(c, err, true,
		"get catalogs", nil) != mysql.Success {
		return nil, err
	}

	log.For(ctx).Info("success get catalogs")
	return catalogs, nil
}

func GetCatalog(c *gin.Context, id int) (*model.Catalog, error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	catalog := &model.Catalog{}
	err := db.First(&catalog, id).Error
	if mysql.ErrorHandleAndLog(c, err, true, "get catalog", id) != mysql.Success {
		return nil, err
	}

	log.For(ctx).Info("success get catalog", zap.Int("catalogId", id))
	return catalog, nil
}

func AddCatalog(c *gin.Context, newCatalog *model.Catalog) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	oldCatalog := model.Catalog{}

	//  check unique
	if !catalogCheckUnique(c, &oldCatalog, newCatalog) {
		return fmt.Errorf("check unique fail")
	}

	err := db.Create(&newCatalog).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"add new catalog", newCatalog.Name) != mysql.Success {
		return err
	}

	log.For(ctx).Info("add new catalog success",
		zap.String("catalogName", newCatalog.Name))
	return nil
}

func UpdateCatalog(c *gin.Context, newCatalog *model.Catalog) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// check if catalog exist
	oldCatalog := model.Catalog{}
	err := db.First(&oldCatalog, newCatalog.ID).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"query catalog", newCatalog.ID) != mysql.Success {
		return err
	}

	//  check unique
	if !catalogCheckUnique(c, &oldCatalog, newCatalog) {
		return fmt.Errorf("check unique fail")
	}

	// update
	err = db.Model(&oldCatalog).Updates(newCatalog).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"update catalog", newCatalog.ID) != mysql.Success {
		return err
	}
	log.For(ctx).Info("update catalog success", zap.String("catalog", newCatalog.Name))

	return nil
}

func catalogCheckUnique(c *gin.Context, oldCatalog *model.Catalog, newCatalog *model.Catalog) bool {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// unique check
	// new value -- old value
	uniqueCheckList := []map[string]mysql.ValuePair{
		{
			"name": mysql.ValuePair{
				NewVal: newCatalog.Name,
				OldVal: oldCatalog.Name,
			},
		},
	}
	for _, checkMap := range uniqueCheckList {
		if !mysql.ShouldUnique(c, ctx, db, checkMap, func(db *gorm.DB) error {
			return db.First(&model.Catalog{}).Error
		}) {
			return false
		}
	}

	return true
}

func DeleteCatalog(c *gin.Context, id int) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	catalog := model.Catalog{}

	// check if catalog exist
	err := db.First(&catalog, id).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"get catalog info", id) != mysql.Success {
		return err
	}

	// todo should check have problem under this catalog

	err = db.Delete(&catalog).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"delete catalog", id) != mysql.Success {
		return err
	}
	log.For(ctx).Info("delete catalog success", zap.Int("catalogId", id))

	return nil
}
