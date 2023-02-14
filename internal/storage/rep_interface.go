package storage

import "context"

type Repositories interface {
	CreateUser(login string, password string, ctx context.Context) (userID int, err error)
	LoginUser(login string, password string, ctx context.Context) (userID int, error error)
	IsUserExistByUserID(userID int, ctx context.Context) (response bool)
	IsUserExistByLogin(login string, ctx context.Context) (response bool)
	InsertDataIntoOrdersTable(order int, userID int, status string, accrual float32, ctx context.Context) (error error)
	IsOrderExistByOrderID(order int, ctx context.Context) (userID int)
	GetAllNewOrders(ctx context.Context) (result []string, err error)
	UpdateOrderInformation(ctx context.Context, orderNum string, status string, accrual float64) (err error)
	GetCurrentBalance(ctx context.Context, userID int) (totalAccrual float64, totalWithdrawal float64, err error)
	InsertDataIntoWithdrawalsTable(order int, userID int, withdrawal float64, ctx context.Context) (error error)
	GetUserOrders(ctx context.Context, userID int) (result []Orders, err error)
	GetUserWithdrawals(ctx context.Context, userID int) (result []Withdrawals, err error)
}
