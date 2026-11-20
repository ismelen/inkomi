package ports

type IPasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) (bool, error)
}
