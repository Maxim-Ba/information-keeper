package domain

type User struct {
	ID             string `json:"id" db:"id"`
	Login          string `json:"login" db:"login"`
	Password       string `json:"password" db:"password"`
	Email          string `json:"email" db:"email"`
	EmailConfirmed bool   `json:"email_confirmed" db:"email_confirmed"`
}
