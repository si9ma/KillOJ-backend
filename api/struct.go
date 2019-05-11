package api

type PageArg struct {
	Page     int    `json:"page" form:"page" binding:"required,min=1"`
	PageSize int    `json:"page_size" form:"page_size" binding:"required,min=1"`
	Order    string `json:"order" form:"order"`
}

type QueryArg struct {
	ID int `json:"id" uri:"id" form:"id" binding:"required"`
}
