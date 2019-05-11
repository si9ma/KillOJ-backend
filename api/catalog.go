package api

import (
	"net/http"

	"github.com/si9ma/KillOJ-backend/srv"

	"github.com/si9ma/KillOJ-backend/auth"
	"github.com/si9ma/KillOJ-backend/middleware"

	"go.uber.org/zap"

	"github.com/si9ma/KillOJ-common/log"

	"github.com/si9ma/KillOJ-backend/wrap"

	"github.com/gin-gonic/gin"

	"github.com/si9ma/KillOJ-common/model"
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
	arg := PageArg{}

	// bind
	if !wrap.ShouldBind(c, &arg, false) {
		return
	}

	catalogs, err := srv.GetAllCatalogs(c, arg.Page, arg.PageSize, arg.Order)
	if err != nil {
		log.For(ctx).Error("get catalogs fail", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, catalogs)
}

func GetCatalog(c *gin.Context) {
	ctx := c.Request.Context()
	arg := QueryArg{}

	// bind
	if !wrap.ShouldBind(c, &arg, true) {
		return
	}

	catalog, err := srv.GetCatalog(c, arg.ID)
	if err != nil {
		log.For(ctx).Error("get catalog info fail", zap.Error(err), zap.Int("catalogId", arg.ID))
		return
	}

	c.JSON(http.StatusOK, catalog)
}

func AddCatalog(c *gin.Context) {
	ctx := c.Request.Context()
	newCatalog := model.Catalog{}

	// bind
	if !wrap.ShouldBind(c, &newCatalog, false) {
		return
	}

	if err := srv.AddCatalog(c, &newCatalog); err != nil {
		log.For(ctx).Error("add catalog fail", zap.Error(err), zap.Int("catalogId", newCatalog.ID))
		return
	}

	c.JSON(http.StatusOK, newCatalog)
}

func UpdateCatalog(c *gin.Context) {
	ctx := c.Request.Context()
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

	// use id in uri path
	newCatalog.ID = uriArg.ID
	if err := srv.UpdateCatalog(c, &newCatalog); err != nil {
		log.For(ctx).Error("update catalog fail", zap.Error(err), zap.Int("catalogId", newCatalog.ID))
		return
	}

	c.JSON(http.StatusOK, newCatalog)
}

func DeleteCatalog(c *gin.Context) {
	ctx := c.Request.Context()
	uriArg := QueryArg{}

	// bind uri params
	if !wrap.ShouldBind(c, &uriArg, true) {
		return
	}

	if err := srv.DeleteCatalog(c, uriArg.ID); err != nil {
		log.For(ctx).Error("delete catalog fail", zap.Error(err), zap.Int("catalogId", uriArg.ID))
		return
	}

	c.JSON(http.StatusOK, nil)
}
