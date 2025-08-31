package dto

type UserRepoDTO struct {
	ID             string
	Login          string
	Email          string
	EmailConfirmed bool
}

type UserAuthReqDTO struct {
	Login    string
	Password string
}
