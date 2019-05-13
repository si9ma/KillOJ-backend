package srv

import (
	"fmt"

	"github.com/jinzhu/gorm"

	"github.com/gin-gonic/gin"
	"github.com/si9ma/KillOJ-backend/auth"
	"github.com/si9ma/KillOJ-backend/gbl"
	"github.com/si9ma/KillOJ-backend/kerror"
	"github.com/si9ma/KillOJ-backend/wrap"
	"github.com/si9ma/KillOJ-common/log"
	"github.com/si9ma/KillOJ-common/model"
	"github.com/si9ma/KillOJ-common/mysql"
	otgrom "github.com/smacker/opentracing-gorm"
	"go.uber.org/zap"
)

func GetAllProblems(c *gin.Context, page, pageSize int, order string) ([]model.Problem, error) {
	var err error
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	myID := auth.GetUserFromJWT(c).ID
	offset := (page - 1) * pageSize
	rawsql := `
		select distinct p.* from problem as p,user as u,user_in_group as up,user_in_contest as uc 
		where u.id = ? and
		(
			p.belong_type = 0 or
			u.id = p.owner_id or
			(p.belong_type = 1 and u.id = up.user_id and p.belong_to_id = up.group_id) or
			(p.belong_type = 2 and u.id = up.user_id and p.belong_to_id = uc.contest_id)
		)
	`
	var problems []model.Problem
	if order != "" {
		err = db.Raw(rawsql, myID).Limit(pageSize).Offset(offset).Order(order).Find(&problems).Error
	} else {
		err = db.Raw(rawsql, myID).Limit(pageSize).Offset(offset).Find(&problems).Error
	}
	// error handle
	if mysql.ErrorHandleAndLog(c, err, true,
		"get problems", nil) != mysql.Success {
		return nil, err
	}

	log.For(ctx).Info("success get problems")
	return problems, nil
}

func GetProblem(c *gin.Context, id int) (*model.Problem, error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	problem := model.Problem{}
	myID := auth.GetUserFromJWT(c).ID

	// get problem
	err := db.First(&problem, id).Error
	if mysql.ErrorHandleAndLog(c, err, true, "get problem", id) != mysql.Success {
		return nil, err
	}

	// if problem is public or problem belong to myself,
	// return
	if problem.BelongType == model.BelongToPublic || problem.OwnerID == myID {
		return &problem, nil
	}

	// otherwise, check permission
	switch problem.BelongType {
	case model.BelongToContest:
		// has permission
		if _, err := GetContest(c, problem.BelongToID); err == nil {
			return &problem, nil
		}
	case model.BelongToGroup:
		// has permission
		if _, err := GetGroup(c, problem.BelongToID); err == nil {
			return &problem, nil
		}
	}

	log.For(ctx).Error("user no permission to access problem", zap.Int("problemID", id))
	wrap.DiscardGinError(c) // discard inner error
	_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
		SetMeta(kerror.ErrNotFound.WithArgs(id))
	return nil, kerror.EmptyError
}

func AddProblem(c *gin.Context, newProblem *model.Problem) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	oldProblem := model.Problem{}

	//  check unique
	if !problemCheckUnique(c, &oldProblem, newProblem) {
		return fmt.Errorf("check unique fail")
	}

	// add new problem
	err := db.Create(&newProblem).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"add new problem", newProblem.Name) != mysql.Success {
		return err
	}

	log.For(ctx).Info("add new problem success",
		zap.String("problemName", newProblem.Name))
	return nil
}

func UpdateProblem(c *gin.Context, newProblem *model.Problem) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// check if problem exist
	oldProblem, err := GetProblem(c, newProblem.ID)
	if err != nil {
		return err
	}

	// check owner
	if err := checkProblemOwner(c, oldProblem); err != nil {
		return err
	}

	//  check unique
	if !problemCheckUnique(c, oldProblem, newProblem) {
		return fmt.Errorf("check unique fail")
	}

	// update
	err = db.Model(oldProblem).Updates(newProblem).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"update problem", newProblem.ID) != mysql.Success {
		return err
	}
	log.For(ctx).Info("update problem success", zap.String("problem", newProblem.Name))

	return nil
}

func problemCheckUnique(c *gin.Context, oldProblem *model.Problem, newProblem *model.Problem) bool {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	// unique check
	// new value -- old value
	uniqueCheckList := []map[string]mysql.ValuePair{
		{
			"name": mysql.ValuePair{
				NewVal: newProblem.Name,
				OldVal: oldProblem.Name,
			},
		},
	}
	for _, checkMap := range uniqueCheckList {
		if !mysql.ShouldUnique(c, ctx, db, checkMap, func(db *gorm.DB) error {
			return db.First(&model.Problem{}).Error
		}) {
			return false
		}
	}

	return true
}

func checkProblemOwner(c *gin.Context, problem *model.Problem) error {
	ctx := c.Request.Context()

	// check owner
	userId := auth.GetUserFromJWT(c).ID
	if userId != problem.OwnerID {
		err := fmt.Errorf("operate problem forbidden")
		log.For(ctx).Error("operate problem fail(forbidden)", zap.Int("problemId", problem.ID))

		_ = c.Error(err).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrForbiddenGeneral)
		return err
	}

	return nil
}

//func DeleteProblem(c *gin.Context, id int) error {
//	ctx := c.Request.Context()
//	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
//
//	// check if problem exist
//	problem, err := GetProblem(c, id)
//	if err != nil {
//		return err
//	}
//
//	// check owner
//	if err := checkProblemOwner(c, problem); err != nil {
//		return err
//	}
//	// todo should check have problem under this problem
//
//	err = db.Delete(&problem).Error
//	if mysql.ErrorHandleAndLog(c, err, true,
//		"delete problem", id) != mysql.Success {
//		return err
//	}
//	log.For(ctx).Info("delete problem success", zap.Int("problemId", id))
//
//	return nil
//}
