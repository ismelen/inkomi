package models

type Role string

const RoleUser Role = "user"

type User struct {
	Id            int
	Email         string
	EmailVerified bool
	PasswordHash  string
	Role          Role
	FCMToken      string
}

func NewPasswordUser(email, passwordHash string) *User {
	return &User{Email: email, PasswordHash: passwordHash, Role: RoleUser}
}

// NewSocialUser keeps the email of account only when the provider verified it.
func NewSocialUser(account ExternalAccount) *User {
	user := &User{Role: RoleUser}
	if account.EmailVerified {
		if email, err := NormalizeEmail(account.Email); err == nil {
			user.Email, user.EmailVerified = email, true
		}
	}
	return user
}

func (u *User) HasPassword() bool {
	return u.PasswordHash != ""
}

// ClaimEmail is called when the owner of the email proved it through a provider. An unverified
// account registered with that email may belong to someone else, so its password is dropped.
func (u *User) ClaimEmail() {
	if !u.EmailVerified {
		u.PasswordHash = ""
	}
	u.EmailVerified = true
}
