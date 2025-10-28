package request

type CreateUser struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	FullName string `json:"full_name"`
}
type UpdateUser struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
	FullName *string `json:"full_name"`
}

type UUIDParam struct {
	UUID string `uri:"uuid" binding:"required,uuid"`
}

type IDParam struct {
	ID int64 `uri:"id" binding:"required,gt=0"`
}
