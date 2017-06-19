package authz

type UserBlockStore interface {
	UserIsBlocked(userUuid string) (bool, error)
}
