package model

type ProblemTestCase struct {
	ID             int    `gorm:"column:id;primary_key" json:"id" binding:"requiredwhenfield=DeleteIt"`
	ProblemID      int    `gorm:"column:problem_id" json:"-"`
	InputData      string `gorm:"column:input_data" json:"input_data" binding:"required"`
	ExpectedOutput string `gorm:"column:expected_output" json:"expected_output" binding:"required"`
	DeleteIt       bool   `gorm:"-" json:"delete_it,omitempty"`
}

// TableName sets the insert table name for this struct type
func (p *ProblemTestCase) TableName() string {
	return "problem_test_case"
}
