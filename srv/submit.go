package srv

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/si9ma/KillOJ-backend/auth"
	"github.com/si9ma/KillOJ-backend/gbl"
	"github.com/si9ma/KillOJ-common/log"
	"github.com/si9ma/KillOJ-common/model"
	"github.com/si9ma/KillOJ-common/mysql"
	otgrom "github.com/smacker/opentracing-gorm"
)

func GetAllSubmitOfGroup(c *gin.Context, id int) (*gorm.DB, error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// no group
	if _, err := GetGroup(c, id); err != nil {
		return nil, err
	}
	return db.Joins("join problem on problem.id = submit.problem_id AND"+
		" problem.belong_type = 1 AND problem.belong_to_id = ?", id), nil
}

func GetAllSubmitOfContest(c *gin.Context, id int, onlyDuringContest bool) (*gorm.DB, error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	contest, err := GetContest(c, id)
	if err != nil {
		// no contest
		return nil, err
	}

	db = db.Joins("join problem on problem.id = submit.problem_id AND"+
		" problem.belong_type = 2 AND problem.belong_to_id = ?", id)

	if onlyDuringContest {
		db = db.Where("created_at BETWEEN '?' AND '?'", contest.StartTime, contest.EndTime)
	}

	return db, nil
}

func GetAllSubmitOfProblem(c *gin.Context, id int) (*gorm.DB, error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// no problem
	if _, err := GetProblem(c, id, false); err != nil {
		return nil, err
	}

	return db.Joins("join problem on problem.id = submit.problem_id AND problem.id = ?", id), nil
}

func GetAllSubmitOfMe(c *gin.Context) (*gorm.DB, error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	myID := auth.GetUserFromJWT(c).ID

	return db.Where("user_id = ?", myID), nil
}

func GetAllSubmitOfUser(c *gin.Context, id int) (*gorm.DB, error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	return db.Where("user_id = ?", id), nil
}

func GetAllSubmit(c *gin.Context, page, pageSize int, order string, of string, id int, onlyDuringContest bool) ([]model.Submit, error) {
	var err error
	var submits []model.Submit

	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	offset := (page - 1) * pageSize

	switch of {
	case "group":
		db, err = GetAllSubmitOfGroup(c, id)
	case "contest":
		db, err = GetAllSubmitOfContest(c, id, onlyDuringContest)
	case "problem":
		db, err = GetAllSubmitOfProblem(c, id)
	case "me":
		db, err = GetAllSubmitOfMe(c)
	case "user":
		db, err = GetAllSubmitOfUser(c, id)
	}

	if err != nil {
		return nil, err
	}

	queryDB := db.Where("is_complete = ?", true).Preload("User").Preload("Problem").Limit(pageSize).Offset(offset)

	if order != "" {
		err = queryDB.Order(order).Find(&submits).Error
	} else {
		err = queryDB.Find(&submits).Error
	}
	// error handle
	if mysql.ErrorHandleAndLog(c, err, true,
		"get submits", nil) != mysql.Success {
		return nil, err
	}

	log.For(ctx).Info("success get submits")
	return submits, nil
}

func GetSubmit(c *gin.Context, id int) (*model.Submit, error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	submit := model.Submit{}
	err := db.Preload("Problem").Preload("User").First(&submit, id).Error

	// error handle
	if mysql.ErrorHandleAndLog(c, err, true,
		"get submit", id) != mysql.Success {
		return nil, err
	}

	// check if user has permission
	if _, err := GetProblem(c, submit.ProblemID, false); err != nil {
		return nil, err
	}

	return &submit, nil
}
