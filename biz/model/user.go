package user

type User struct {
	NickName string

	// binds
	Email string
	Phone string
	CID   string
	Tags  map[string]struct{}
}
