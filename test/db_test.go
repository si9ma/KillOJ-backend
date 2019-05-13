package model

import (
	"testing"

	"github.com/si9ma/KillOJ-common/model"

	"github.com/si9ma/KillOJ-common/kjson"

	"github.com/jinzhu/gorm"
	"github.com/si9ma/KillOJ-common/mysql"
)

func TestBelongsTo(t *testing.T) {
	var (
		db  *gorm.DB
		err error
	)

	if db, err = mysql.GetTestDB(); err != nil {
		t.Fatal("init mysql fail", err)
	}

	var (
		user  model.User
		group model.Group
	)

	db.First(&group)
	db.Model(&group).Related(&user, "OwnerID")
	t.Log(kjson.MarshalString(user))
	t.Log(kjson.MarshalString(group))

	var (
		problem1 model.Problem
		submit1  model.Submit
	)

	db.First(&submit1)
	db.Model(&submit1).Related(&problem1)
	t.Log(kjson.MarshalString(problem1))
	t.Log(kjson.MarshalString(submit1))
}

func TestHaveMany(t *testing.T) {
	var (
		db  *gorm.DB
		err error
	)

	if db, err = mysql.GetTestDB(); err != nil {
		t.Fatal("init mysql fail", err)
	}

	var (
		problem         model.Problem
		problemTestCase []model.ProblemTestCase
	)

	db.First(&problem)
	db.Model(&problem).Related(&problemTestCase)
	t.Log(kjson.MarshalString(problem))
	t.Log(kjson.MarshalString(problemTestCase))
}

type Userr struct {
	gorm.Model
	CreditCards []CreditCard `gorm:"foreignkey:UserRefer"`
}

type CreditCard struct {
	gorm.Model
	Number    string
	UserRefer uint
}

func TestDB(t *testing.T) {
	db, err := mysql.GetTestDB()
	db.LogMode(true)
	if err != nil {
		t.Fatal(err)
	}
	user := Userr{}
	creditCards := []CreditCard{}
	db.First(&user)
	db.Model(&user).Related(&creditCards, "UserRefer")

	t.Log(kjson.MarshalString(&user))
	t.Log(kjson.MarshalString(&creditCards))

	problem := model.Problem{}
	db.First(&problem)
	t.Log(kjson.MarshalString(&problem))

	rawsql := `
select distinct p.* from problem as p,user as u,user_in_group as up,user_in_contest as uc 
    where u.id = ? and
    (
		p.belong_type = 0 or
		(p.belong_type = 1 and u.id = up.user_id and p.belong_to_id = up.group_id) or
		(p.belong_type = 2 and u.id = up.user_id and p.belong_to_id = uc.contest_id)
    )
`
	var problems []model.Problem
	db.Raw(rawsql, 30).Limit(10).Offset(0).Find(&problems)
	t.Log(kjson.MarshalString(&problems))

	err = db.Raw("delete from problem_has_tag as t where t.tag_id = 22 and t.problem_id = 32").Error
	err = db.First(&problem, 100).Error
	t.Log(err)
}
