package srv

import (
	"fmt"
	"strconv"
	"time"

	"github.com/si9ma/KillOJ-common/kjson"

	"github.com/si9ma/KillOJ-common/constants"

	"github.com/si9ma/KillOJ-common/kredis"

	"github.com/si9ma/KillOJ-common/judge"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/opentracing/opentracing-go"

	"github.com/si9ma/KillOJ-backend/data"

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

const (
	noOfSql = `
		select distinct p.* from problem as p,user_in_group as up,user_in_contest as uc 
		where 
			p.belong_type = 0 or
			p.owner_id = ? or
			(p.belong_type = 1 and up.user_id = ? and p.belong_to_id = up.group_id) or
			(p.belong_type = 2 and uc.user_id = ? and p.belong_to_id = uc.contest_id)
		`
	ofTagSql = `
		select distinct p.* from problem as p,user_in_group as up,user_in_contest as uc,problem_has_tag as pt 
		where
		pt.problem_id = p.id and
		pt.tag_id = ? and
		(
			p.belong_type = 0 or
			p.owner_id = ? or
			(p.belong_type = 1 and up.user_id = ? and p.belong_to_id = up.group_id) or
			(p.belong_type = 2 and uc.user_id = ? and p.belong_to_id = uc.contest_id)
		)
		`
)

func GetAllProblemsOfGroup(c *gin.Context, id int) (*gorm.DB, error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	_, err := GetGroup(c, id)
	if err != nil {
		return nil, err
	}

	return db.Joins("join `group` on `group`.id = problem.belong_to_id and problem.belong_type = 1"), nil
}

func GetAllProblemsOfContest(c *gin.Context, id int) (*gorm.DB, error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	_, err := GetContest(c, id)
	if err != nil {
		return nil, err
	}

	return db.Joins("join contest on contest.id = problem.belong_to_id and problem.belong_type = 2"), nil
}

func GetAllProblemsOfTag(c *gin.Context, id int) *gorm.DB {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	myID := auth.GetUserFromJWT(c).ID

	return db.Raw(ofTagSql, id, myID, myID, myID)
}

func GetAllProblemsOf(c *gin.Context) *gorm.DB {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	myID := auth.GetUserFromJWT(c).ID

	return db.Raw(noOfSql, myID, myID, myID)
}

func GetAllProblems(c *gin.Context, page, pageSize int, order string, of string, id int) ([]model.Problem, error) {
	var err error
	var problems []model.Problem
	var db *gorm.DB

	ctx := c.Request.Context()
	offset := (page - 1) * pageSize

	switch of {
	case "group":
		db, err = GetAllProblemsOfGroup(c, id)
	case "contest":
		db, err = GetAllProblemsOfContest(c, id)
	case "tag":
		db = GetAllProblemsOfTag(c, id)
	default:
		db = GetAllProblemsOf(c)
	}

	if err != nil {
		return nil, err
	}

	queryDB := db.Preload("Tags").Preload("UpVoteUsers", "attitude = ?", model.Up).
		Preload("DownVoteUsers", "attitude = ?", model.Down).Preload("Catalog").
		Limit(pageSize).Offset(offset)
	if order != "" {
		err = queryDB.Order(order).Find(&problems).Error
	} else {
		err = queryDB.Find(&problems).Error
	}
	// error handle
	if mysql.ErrorHandleAndLog(c, err, true,
		"get problems", nil) != mysql.Success {
		return nil, err
	}

	log.For(ctx).Info("success get problems")
	return problems, nil
}

func GetProblem(c *gin.Context, id int, forUpdate bool) (*model.Problem, error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	problem := model.Problem{}
	myID := auth.GetUserFromJWT(c).ID

	// get problem
	queryDB := db.Preload("Tags").Preload("ProblemSamples").Preload("ProblemTestCases").Preload("Owner").
		Preload("UpVoteUsers", "attitude = ?", model.Up).
		Preload("DownVoteUsers", "attitude = ?", model.Down).
		Preload("Catalog")

	submit := model.Submit{}
	err := db.First(&submit).Error
	if r := mysql.ErrorHandleAndLog(c, err, false,
		"check if user have success submit for problem", id); r == mysql.Success {
		// if user haven't has any success submit, we shouldn't return comments
		queryDB = queryDB.Preload("Comments").Preload("Comments.From").Preload("Comments.To")
	} else if r == mysql.DB_ERROR {
		return nil, err
	}

	err = queryDB.First(&problem, id).Error
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

	// when set forupdate flag
	isOwner := false
	if forUpdate {
		// if user is the owner of problem,
		// return test case
		if err := checkProblemOwner(c, &problem); err == nil {
			isOwner = true
		}
	}
	if !isOwner {
		// if user not owner , clear test case
		problem.ProblemTestCases = []model.ProblemTestCase{}
	}

	log.For(ctx).Error("user no permission to access problem", zap.Int("problemID", id))
	wrap.DiscardGinError(c) // discard inner error
	_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
		SetMeta(kerror.ErrNotFound.WithArgs(id))
	return nil, kerror.EmptyError
}

// clear all id of tag,
// avoid auto update by gorm
func clearTagIDs(problem *model.Problem) {
	for i := range problem.Tags {
		problem.Tags[i].ID = 0
	}
}

// clear all id of sample,
// avoid auto update by gorm
func clearSampleIDs(problem *model.Problem) {
	for i := range problem.ProblemSamples {
		problem.ProblemSamples[i].ID = 0
	}
}

// clear all id of test case,
// avoid auto update by gorm
func clearTestCaseIDs(problem *model.Problem) {
	for i := range problem.ProblemTestCases {
		problem.ProblemTestCases[i].ID = 0
	}
}

// filter tag, if tag already exist, we will not insert it
func filterTags(c *gin.Context, problem *model.Problem) error {
	var tagNames []string
	var inDbTags []model.Tag

	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	for _, tag := range problem.Tags {
		tagNames = append(tagNames, tag.Name)
	}

	err := db.Where("name in (?)", tagNames).Find(&inDbTags).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"check if tags exist", tagNames) != mysql.Success {
		return err
	}

	helpMap := make(map[string]int)
	for i, tag := range problem.Tags {
		helpMap[tag.Name] = i
	}
	// replace already exist
	for _, tag := range inDbTags {
		if v, ok := helpMap[tag.Name]; ok {
			// replace
			problem.Tags[v] = tag
		}
	}

	return nil
}

func AddProblem(c *gin.Context, newProblem *model.Problem) error {
	ctx := c.Request.Context()
	dbWithAutoUpdate := gbl.DB.Set("gorm:association_autoupdate", true).Set("gorm:association_autocreate", true)
	db := otgrom.SetSpanToGorm(ctx, dbWithAutoUpdate)
	oldProblem := model.Problem{}

	//  check unique
	if !problemCheckUnique(c, &oldProblem, newProblem) {
		return fmt.Errorf("check unique fail")
	}

	// must clear ids
	clearTagIDs(newProblem)
	clearSampleIDs(newProblem)
	clearTestCaseIDs(newProblem)

	// filter tags
	if err := filterTags(c, newProblem); err != nil {
		return err
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

func deleteTags(c *gin.Context, problem *model.Problem) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	var deleteTagIDs []int
	var remainTags []model.Tag

	for _, tag := range problem.Tags {
		if tag.DeleteIt {
			deleteTagIDs = append(deleteTagIDs, tag.ID)
		} else {
			remainTags = append(remainTags, tag)
		}
	}

	// no tag need be deleted
	if len(deleteTagIDs) == 0 {
		log.For(ctx).Info("no tag need delete", zap.Int("problemID", problem.ID))
		return nil
	}

	err := db.Where("tag_id in (?) AND problem_id = ?", deleteTagIDs, problem.ID).Delete(&model.ProblemHasTag{}).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"delete tags for problem", problem.ID) != mysql.Success {
		return err
	}

	problem.Tags = remainTags
	return nil
}

func deleteSamples(c *gin.Context, problem *model.Problem) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	var deleteSamplesID []int
	var remainSamples []model.ProblemSample

	for _, sample := range problem.ProblemSamples {
		if sample.DeleteIt {
			deleteSamplesID = append(deleteSamplesID, sample.ID)
		} else {
			remainSamples = append(remainSamples, sample)
		}
	}

	if len(deleteSamplesID) == 0 {
		log.For(ctx).Info("no sample need delete", zap.Int("problemID", problem.ID))
		return nil
	}

	err := db.Where("id in (?) AND problem_id = ?", deleteSamplesID, problem.ID).Delete(&model.ProblemSample{}).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"delete samples for problem", problem.ID) != mysql.Success {
		return err
	}

	problem.ProblemSamples = remainSamples
	return nil
}

func checkSamples(c *gin.Context, problem *model.Problem) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	var sampleIDs []int

	for _, sample := range problem.ProblemSamples {
		if sample.ID != 0 {
			// if sample id is 0 --> add sample
			sampleIDs = append(sampleIDs, sample.ID)
		}
	}

	tmpProblem := model.Problem{
		ID: problem.ID,
	}
	err := db.Preload("ProblemSamples", "id in (?)", sampleIDs).First(&tmpProblem).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"find samples of problem", problem.ID) != mysql.Success {
		return err
	}

	var notExistSamples []int
	if len(sampleIDs) != len(tmpProblem.ProblemSamples) {
		// some samples not belong to this problem
		helpMap := make(map[int]int)
		for i, sample := range tmpProblem.ProblemSamples {
			helpMap[sample.ID] = i
		}
		for _, id := range sampleIDs {
			if _, ok := helpMap[id]; !ok {
				notExistSamples = append(notExistSamples, id)
			}
		}
		fields := map[string]interface{}{
			"samples": notExistSamples,
		}
		log.For(ctx).Error("these samples no exist", zap.Any("samples", sampleIDs))

		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrNotExist.WithArgs(notExistSamples).With(fields))
		return fmt.Errorf("some samples no exist")
	}

	return nil
}

func deleteTestCases(c *gin.Context, problem *model.Problem) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	var deleteTestCasesID []int
	var remainTestCases []model.ProblemTestCase

	for _, testCase := range problem.ProblemTestCases {
		if testCase.DeleteIt {
			deleteTestCasesID = append(deleteTestCasesID, testCase.ID)
		} else {
			remainTestCases = append(remainTestCases, testCase)
		}
	}

	if len(deleteTestCasesID) == 0 {
		log.For(ctx).Info("no testCase need delete", zap.Int("problemID", problem.ID))
		return nil
	}

	err := db.Where("id in (?) AND problem_id = ?", deleteTestCasesID, problem.ID).Delete(&model.ProblemTestCase{}).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"delete testCases for problem", problem.ID) != mysql.Success {
		return err
	}

	problem.ProblemTestCases = remainTestCases
	return nil
}

func atLeastOneTestCase(c *gin.Context, problem *model.Problem) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)

	addHowManyCase := 0
	for _, testCase := range problem.ProblemTestCases {
		// delete it
		if testCase.DeleteIt {
			addHowManyCase--
			continue
		}

		// add one
		if testCase.ID == 0 {
			addHowManyCase++
		}
	}

	casesInDB := 0
	err := db.Table("problem_test_case").Where("problem_id = ?", problem.ID).Count(&casesInDB).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"check how many test case for problem", problem.ID) != mysql.Success {
		return err
	}

	if casesInDB+addHowManyCase <= 0 {
		log.For(ctx).Error("one problem at least have one test case")

		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrAtLeast.WithArgs(1, "test case"))
		return kerror.EmptyError
	}

	return nil
}

func checkTestCases(c *gin.Context, problem *model.Problem) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	var testCaseIDs []int

	for _, testCase := range problem.ProblemTestCases {
		if testCase.ID != 0 {
			// if testCase id is 0 --> add testCase
			testCaseIDs = append(testCaseIDs, testCase.ID)
		}
	}

	tmpProblem := model.Problem{
		ID: problem.ID,
	}
	err := db.Preload("ProblemTestCases", "id in (?)", testCaseIDs).First(&tmpProblem).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"find testCases of problem", problem.ID) != mysql.Success {
		return err
	}

	var notExistTestCases []int
	if len(testCaseIDs) != len(tmpProblem.ProblemTestCases) {
		// some testCases not belong to this problem
		helpMap := make(map[int]int)
		for i, testCase := range tmpProblem.ProblemTestCases {
			helpMap[testCase.ID] = i
		}
		for _, id := range testCaseIDs {
			if _, ok := helpMap[id]; !ok {
				notExistTestCases = append(notExistTestCases, id)
			}
		}
		fields := map[string]interface{}{
			"testCases": notExistTestCases,
		}
		log.For(ctx).Error("these testCases no exist", zap.Any("testCases", testCaseIDs))

		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrNotExist.WithArgs(notExistTestCases).With(fields))
		return fmt.Errorf("some testCases no exist")
	}

	return nil
}

func UpdateProblem(c *gin.Context, newProblem *model.Problem) error {
	ctx := c.Request.Context()
	dbWithAutoUpdate := gbl.DB.Set("gorm:association_autoupdate", true).Set("gorm:association_autocreate", true)
	db := otgrom.SetSpanToGorm(ctx, dbWithAutoUpdate)

	// check if problem exist
	oldProblem, err := GetProblem(c, newProblem.ID, false)
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

	// check if all sample exist
	if err := checkSamples(c, newProblem); err != nil {
		return err
	}

	// check if all test case exist
	if err := checkTestCases(c, newProblem); err != nil {
		return err
	}

	// at least one test case check
	if err := atLeastOneTestCase(c, newProblem); err != nil {
		return err
	}

	// delete tags which be mark as delete
	if err := deleteTags(c, newProblem); err != nil {
		return err
	}

	// delete samples which be mark as delete
	if err := deleteSamples(c, newProblem); err != nil {
		return err
	}

	// delete test cases which be mark as delete
	if err := deleteTestCases(c, newProblem); err != nil {
		return err
	}

	// must clear ids
	clearTagIDs(newProblem)

	// filter tags
	if err := filterTags(c, newProblem); err != nil {
		return err
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

// vote for problem
func VoteProblem(c *gin.Context, id int, attitude int) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	myID := auth.GetUserFromJWT(c).ID

	// check if problem exist
	if _, err := GetProblem(c, id, false); err != nil {
		return err
	}

	vote := model.UserVoteProblem{
		UserID:    myID,
		ProblemID: id,
		Attitude:  attitude,
	}

	err := db.Save(&vote).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"user vote for problem", id) != mysql.Success {
		return err
	}

	return nil
}

// submit answer
func Submit(c *gin.Context, submitArg *data.SubmitArg) error {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	redisCli := kredis.WrapRedisClusterClient(ctx, gbl.Redis)
	myID := auth.GetUserFromJWT(c).ID

	// check if problem exist
	if _, err := GetProblem(c, submitArg.ProblemID, false); err != nil {
		return err
	}

	// check if user have running task
	k := constants.UserProblemSubmitIsCompletePrefix + strconv.Itoa(myID) + "_" + strconv.Itoa(submitArg.ProblemID)
	err := redisCli.Get(k).Err()
	res := kredis.ErrorHandleAndLog(c, err, false,
		"check if user has running submit", k, nil)
	switch res {
	case kredis.Success:
		log.For(ctx).Error("user already have running submit", zap.Int("problemID", submitArg.ProblemID))
		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).SetMeta(kerror.ErrHaveRunningTask)
		return kerror.EmptyError
	case kredis.DB_ERROR:
		return err
	case kredis.NotFound:
		break // continue
	}

	submit := model.Submit{
		ProblemID:  submitArg.ProblemID,
		UserID:     myID,
		SourceCode: submitArg.SourceCode,
		Language:   submitArg.Language,
	}

	err = db.Save(&submit).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"save user submit", submitArg.ProblemID) != mysql.Success {
		return err
	}

	// submit to judger
	if err := submit2Judger(c, submit.ID); err != nil {
		return err
	}

	// save redis
	err = redisCli.Set(k, false, time.Hour).Err()
	if res := kredis.ErrorHandleAndLog(c, err, true,
		"save user submit to redis", k, nil); res != kredis.Success {
		return err
	}

	return nil
}

// submit to judger
func submit2Judger(c *gin.Context, submitID int) error {
	bgCtx := c.Request.Context()
	span, ctx := opentracing.StartSpanFromContext(bgCtx, "sendTask")
	defer span.Finish()

	judgeTask := tasks.Signature{
		Name: "judge",
		Args: []tasks.Arg{
			{
				Name:  "submitId",
				Type:  "int",
				Value: submitID,
			},
		},
	}

	if _, err := gbl.MachineryServer.SendTaskWithContext(ctx, &judgeTask); err != nil {
		log.For(ctx).Error("send async job fail", zap.Error(err), zap.Int("submitID", submitID))
		wrap.SetInternalServerError(c, err)
		return err
	}
	log.For(ctx).Info("send async jos success", zap.Int("submitID", submitID))

	return nil
}

// get last submit,
// is set success flag,
// return lease successful submit
func GetLastSubmit(c *gin.Context, id int, needSuccess bool, needComplete bool) (*model.Submit, error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	myID := auth.GetUserFromJWT(c).ID

	// check if problem exist
	if _, err := GetProblem(c, id, false); err != nil {
		return nil, err
	}

	submit := model.Submit{}
	queryDB := db.Where("problem_id = ?", id)
	if needComplete {
		queryDB = queryDB.Where("is_complete = ?", true)
	}
	if needSuccess {
		// query last successful result
		queryDB = queryDB.Where("result = ?", judge.AcceptedStatus.Code)
	}
	err := queryDB.Where("user_id = ?", myID).Last(&submit).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"get last user submit", "last submit") != mysql.Success {
		return nil, err
	}

	return &submit, nil
}

// get result
func GetResult(c *gin.Context, problemID int) (*judge.OuterResult, error) {
	ctx := c.Request.Context()
	redisCli := kredis.WrapRedisClusterClient(ctx, gbl.Redis)
	myID := auth.GetUserFromJWT(c).ID

	// check if problem exist
	if _, err := GetProblem(c, problemID, false); err != nil {
		return nil, err
	}

	// check if user have task
	isCompleteKey := constants.UserProblemSubmitIsCompletePrefix + strconv.Itoa(myID) + "_" + strconv.Itoa(problemID)
	res, err := redisCli.Get(isCompleteKey).Result()
	r := kredis.ErrorHandleAndLog(c, err, false,
		"check if user has running submit", isCompleteKey, nil)
	switch r {
	case kredis.Success:
		break // continue
	case kredis.DB_ERROR:
		return nil, err
	case kredis.NotFound:
		log.For(ctx).Error("no result for problem", zap.Int("problemID", problemID))
		arg := fmt.Sprintf("submit for %d", problemID)
		_ = c.Error(err).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrNotFound.WithArgs(arg))
		return nil, err
	}

	isComplete := false
	if isComplete, err = strconv.ParseBool(res); err != nil {
		log.For(ctx).Error("parse is complete redis result fail", zap.Error(err))
		wrap.SetInternalServerError(c, err)
		return nil, err
	}

	// task haven't complete
	if !isComplete {
		log.For(ctx).Info("submit haven't complete", zap.Int("problemID", problemID))
		_ = c.Error(kerror.EmptyError).SetType(gin.ErrorTypePublic).
			SetMeta(kerror.ErrNotComplete)
		return nil, kerror.EmptyError
	}

	// get last submit
	submit, err := GetLastSubmit(c, problemID, false, true)
	if err != nil {
		wrap.DiscardGinError(c)
		wrap.SetInternalServerError(c, err)
		return nil, err
	}

	// get result
	resultKey := constants.SubmitStatusKeyPrefix + strconv.Itoa(submit.ID)
	res, err = redisCli.Get(resultKey).Result()
	r = kredis.ErrorHandleAndLog(c, err, false,
		"get run result of problem", resultKey, submit.ID)
	if r != kredis.Success {
		wrap.SetInternalServerError(c, err)
		return nil, err
	}

	result := judge.OuterResult{}
	if err := kjson.UnmarshalString(res, &result); err != nil {
		log.For(ctx).Error("unmarshal result to result fail", zap.Error(err), zap.Int("problemID", problemID))
		wrap.SetInternalServerError(c, err)
		return nil, err
	}

	// return result and delete redis key
	err = redisCli.Del(isCompleteKey).Err()
	if kredis.ErrorHandleAndLog(c, err, true,
		"remove the run result of submit", isCompleteKey, nil) != kredis.Success {
		return nil, err
	}
	err = redisCli.Del(resultKey).Err()
	if kredis.ErrorHandleAndLog(c, err, true,
		"remove the run result of submit", resultKey, nil) != kredis.Success {
		return nil, err
	}

	return &result, nil
}

func Comment4Problem(c *gin.Context, commentArg *data.CommentArg) (*model.Comment, error) {
	ctx := c.Request.Context()
	db := otgrom.SetSpanToGorm(ctx, gbl.DB)
	myID := auth.GetUserFromJWT(c).ID

	// check if problem exist
	if _, err := GetProblem(c, commentArg.ProblemID, false); err != nil {
		return nil, err
	}

	// for reply
	if commentArg.ForComment > 0 {
		// check if this comment available
		comment := model.Comment{}
		err := db.Where("id = ? AND from_id = ? ", commentArg.ForComment, commentArg.ToID).
			Or("for_comment = ? AND from_id = ?", commentArg.ForComment, commentArg.ToID).
			First(&comment).Error
		if mysql.ErrorHandleAndLog(c, err, true,
			"check if comment exist", commentArg.ForComment) != mysql.Success {
			return nil, err
		}
	}

	comment := model.Comment{
		ProblemID:  commentArg.ProblemID,
		FromID:     myID,
		ToID:       commentArg.ToID,
		Content:    commentArg.Content,
		ForComment: commentArg.ForComment,
	}

	err := db.Save(&comment).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"save comment", commentArg.ForComment) != mysql.Success {
		return nil, err
	}

	// query user info
	err = db.Preload("From").Preload("To").First(&comment, comment.ID).Error
	if mysql.ErrorHandleAndLog(c, err, true,
		"query comment info after add comment", commentArg.ForComment) != mysql.Success {
		return nil, err
	}

	return &comment, nil
}
