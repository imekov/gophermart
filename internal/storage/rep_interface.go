package storage

import "context"

type Repositories interface {
	CreateUser(login string, password string, ctx context.Context) (userID int, err error)
	LoginUser(login string, password string, ctx context.Context) (userID int, error error)
	IsUserExistByUserID(userID int, ctx context.Context) (response bool)
	IsUserExistByLogin(login string, ctx context.Context) (response bool)
}
