package api

import (
	"net/http"

	"github.com/jinzhu/gorm"

	"github.com/si9ma/KillOJ-backend/auth"
	"github.com/si9ma/KillOJ-backend/middleware"

	"go.uber.org/zap"

	"github.com/si9ma/KillOJ-common/log"
	"github.com/si9ma/KillOJ-common/mysql"

	"github.com/si9ma/KillOJ-backend/wrap"

	"github.com/si9ma/KillOJ-backend/gbl"

	"github.com/gin-gonic/gin"

	"github.com/si9ma/KillOJ-common/model"
	otgrom "github.com/smacker/opentracing-gorm"
)

func setupCatalog(r *gin.Engine) {
	// everyone can access
	r.GET("/catalogs", GetAllCatalogs)
	r.GET("/catalogs/:id", GetCatalog)

	// need auth
	auth.AuthGroup.POST("/catalogs",
		middleware.AuthorizateFunc(AddCatalog, auth.Administrator, auth.Maintainer))
	auth.AuthGroup.PUT("/catalogs/:id",
		middleware.AuthorizateFunc(UpdateCatalog, auth.Administrator, auth.Maintainer))
	auth.AuthGroup.DELETE("/catalogs/:id",
		middleware.AuthorizateFunc(DeleteCatalog, auth.Administrator, auth.Maintainer))
}

func GetAllCatalogs(c *gin.Context) {
	var err error

	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	arg := PageArg{}

	// bind
	if !wrap.ShouldBind(c, &arg, false) {
		return
	}
	offset := (arg.Page - 1) * arg.PageSize

	catalogs := []*model.Catalog{}

	if arg.Order != "" {
		err = db.Model(&model.Catalog{}).Order(arg.Order).Offset(offset).Limit(arg.PageSize).Find(&catalogs).Error
	} else {
		err = db.Model(&model.Catalog{}).Offset(offset).Limit(arg.PageSize).Find(&catalogs).Error
	}

	// error handle
	if mysql.ErrorHandleAndLog(c, err, true,
		"get catalogs", nil) != mysql.Success {
		return
	}

	log.For(ctx).Info("success get catalogs")
	c.JSON(http.StatusOK, catalogs)
}

func GetCatalog(c *gin.Context) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	arg := QueryArg{}

	// bind
	if !wrap.ShouldBind(c, &arg, true) {
		return
	}

	catalog := &model.Catalog{}
	err := db.First(&catalog, arg.ID).Error
	if mysql.ErrorHandleAndLog(c, err, true, "get catalog", arg.ID) != mysql.Success {
		return
	}

	log.For(ctx).Info("success get catalog", zap.Int("catalogId", arg.ID))
	c.JSON(http.StatusOK, catalog)
}

func AddCatalog(c *gin.Context) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	newCatalog, oldCatalog := model.Catalog{}, model.Catalog{}

	// bind
	if !wrap.ShouldBind(c, &newCatalog, false) {
		return
	}

	//  check unique
	if !checkUnique(c, oldCatalog, newCatalog) {
		return
	}

	err := db.Create(&newCatalog).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"add new catalog", newCatalog.Name) != mysql.Success {
		return
	}

	log.For(ctx).Info("add new catalog success",
		zap.String("catalogName", newCatalog.Name))
	c.JSON(http.StatusOK, newCatalog)
}

func UpdateCatalog(c *gin.Context) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	newCatalog := model.Catalog{}
	uriArg := QueryArg{}

	// bind uri params
	if !wrap.ShouldBind(c, &uriArg, true) {
		return
	}

	// bind request params
	if !wrap.ShouldBind(c, &newCatalog, false) {
		return
	}

	// check if catalog exist
	oldCatalog := model.Catalog{}
	err := db.First(&oldCatalog, uriArg.ID).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"query catalog", uriArg.ID) != mysql.Success {
		return
	}

	//  check unique
	if !checkUnique(c, oldCatalog, newCatalog) {
		return
	}

	// these field shouldn't update
	newCatalog.ID = oldCatalog.ID

	// update
	err = db.Model(&oldCatalog).Updates(newCatalog).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"update catalog", newCatalog.ID) != mysql.Success {
		return
	}
	if err := db.Model(&oldCatalog).Updates(newCatalog).Error; err != nil {
		log.For(ctx).Error("update  fail", zap.Error(err))
		_ = c.Error(err).SetType(gin.ErrorTypePrivate)
		return
	}
	log.For(ctx).Info("update catalog success", zap.String("catalog", newCatalog.Name))

	c.JSON(http.StatusOK, newCatalog)
}

func checkUnique(c *gin.Context, oldCatalog model.Catalog, newCatalog model.Catalog) bool {
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

func DeleteCatalog(c *gin.Context) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	uriArg := QueryArg{}
	catalog := model.Catalog{}

	// bind uri params
	if !wrap.ShouldBind(c, &uriArg, true) {
		return
	}

	// check if catalog exist
	err := db.First(&catalog, uriArg.ID).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"get catalog info", uriArg.ID) != mysql.Success {
		return
	}

	// todo should check have problem under this catalog

	err = db.Delete(&catalog).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"delete catalog", uriArg.ID) != mysql.Success {
		return
	}
	log.For(ctx).Info("delete catalog success", zap.Int("catalogId", uriArg.ID))

	c.JSON(http.StatusOK, nil)
}
