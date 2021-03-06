package api

type PageArg struct {
	Page     int    `json:"page" form:"page" binding:"required,min=1"`
	PageSize int    `json:"page_size" form:"page_size" binding:"required,min=1"`
	Order    string `json:"order" form:"order"`
	Of       string `json:"of" form:"of" binding:"omitempty,oneof=group tag contest"`
	ID       int    `json:"id" form:"id" binding:"requiredwhenfield=Of"`
}

type SubmitGetArg struct {
	Page              int    `json:"page" form:"page" binding:"required,min=1"`
	PageSize          int    `json:"page_size" form:"page_size" binding:"required,min=1"`
	Order             string `json:"order" form:"order"`
	Of                string `json:"of" form:"of" binding:"omitempty,oneof=group contest problem me user"`
	ID                int    `json:"id" form:"id" binding:"requiredwhenfield=Of"`
	OnlyDuringContest bool   `json:"only_during_contest"`
}

type QueryArg struct {
	ID int `json:"id" uri:"id" form:"id" binding:"required"`
}

type uuidArg struct {
	UUID string `uri:"uuid" binding:"uuid,required"`
}

type joinArg struct {
	Password string `json:"password"`
}

type QueryExtraArg struct {
	ForUpdate bool `json:"for_update" form:"for_update"`
}

type getSubmitArg struct {
	Success bool `json:"success" form:"success"`
}
