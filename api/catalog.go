package api

//
//import (
//	"net/http"
//
//	"github.com/si9ma/KillOJ-common/log"
//	"github.com/si9ma/KillOJ-common/mysql"
//	"go.uber.org/zap"
//
//	"github.com/si9ma/KillOJ-backend/wrap"
//
//	"github.com/si9ma/KillOJ-backend/gbl"
//
//	"github.com/gin-gonic/gin"
//
//	"github.com/julienschmidt/httprouter"
//	"github.com/si9ma/KillOJ-common/model"
//	otgrom "github.com/smacker/opentracing-gorm"
//	"github.com/smallnest/gen/dbmeta"
//)
//
//func setupCatalog(r *gin.Engine) {
//	r.GET("/catalogs", GetAllCatalogs)
//	r.POST("/catalogs", AddCatalog)
//	r.GET("/catalogs/:id", GetCatalog)
//	r.PUT("/catalogs/:id", UpdateCatalog)
//	r.DELETE("/catalogs/:id", DeleteCatalog)
//}
//
//func GetAllCatalogs(c *gin.Context) {
//	var err error
//
//	ctx := c.Request.Context()
//	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
//	arg := PageArg{}
//
//	// bind
//	if !wrap.ShouldBind(c, &arg) {
//		return
//	}
//	offset := (arg.Page - 1) * arg.PageSize
//
//	catalogs := []*model.Catalog{}
//
//	if arg.Order != "" {
//		err = db.Model(&model.Catalog{}).Order(arg.Order).Offset(offset).Limit(arg.PageSize).Find(&catalogs).Error
//	} else {
//		err = db.Model(&model.Catalog{}).Offset(offset).Limit(arg.PageSize).Find(&catalogs).Error
//	}
//	if hasErr, _ := mysql.ApplyDBError(c, err); hasErr {
//		log.For(ctx).Error("get all catalogs fail", zap.Error(err))
//		return
//	}
//
//	log.For(ctx).Info("success get catalogs")
//	c.JSON(http.StatusOK, catalogs)
//}
//
//func GetCatalog(c *gin.Context) {
//	ctx := c.Request.Context()
//	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
//
//	id := ps.ByName("id")
//	catalog := &model.Catalog{}
//	if DB.First(catalog, id).Error != nil {
//		http.NotFound(w, r)
//		return
//	}
//	writeJSON(w, catalog)
//}
//
//func AddCatalog(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
//	catalog := &model.Catalog{}
//
//	if err := readJSON(r, catalog); err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//	if err := DB.Save(catalog).Error; err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//	writeJSON(w, catalog)
//}
//
//func UpdateCatalog(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
//	id := ps.ByName("id")
//
//	catalog := &model.Catalog{}
//	if DB.First(catalog, id).Error != nil {
//		http.NotFound(w, r)
//		return
//	}
//
//	updated := &model.Catalog{}
//	if err := readJSON(r, updated); err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//
//	if err := dbmeta.Copy(catalog, updated); err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//
//	if err := DB.Save(catalog).Error; err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//	writeJSON(w, catalog)
//}
//
//func DeleteCatalog(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
//	id := ps.ByName("id")
//	catalog := &model.Catalog{}
//
//	if DB.First(catalog, id).Error != nil {
//		http.NotFound(w, r)
//		return
//	}
//	if err := DB.Delete(catalog).Error; err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//	w.WriteHeader(http.StatusOK)
//}
