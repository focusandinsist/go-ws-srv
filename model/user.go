package model

type User struct {
	UID  int64
	NickName string

	// binds
	Email string
	Phone string
	Tags  map[string]struct{}
}
