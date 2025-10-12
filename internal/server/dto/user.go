package dto

// UserRepoDTO struct
type UserRepoDTO struct {
	ID             string
	Login          string
	Email          string
	EmailConfirmed bool
}


// UserAuthReqDTO struct
type UserAuthReqDTO struct {
	Login    string
	Password string
}
