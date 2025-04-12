package domain

type User struct {
	ID           string `db:"id"`
	Email        string `db:"email"`
	PasswordHash string `db:"password_hash"`
	FirstName    string `db:"first_name"`
	LastName     string `db:"last_name"`
}
