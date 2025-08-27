package domain

type User struct {
	ID             string
	Login          string
	Password       string
	Email          string
	EmailConfirmed bool
}
