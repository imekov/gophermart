package storage

import (
	"context"
	"log"
)

func (s PostgreConnect) GetCurrentBalance(ctx context.Context, userID int) (totalAccrual float64, totalWithdrawal float64, err error) {

	tx, err := s.DBConnect.BeginTx(ctx, nil)
	if err != nil {
		log.Print(err)
		return 0, 0, err
	}

	err = tx.QueryRow("select balance, total_withdrawal from current_balance where user_id = $1;", userID).Scan(&totalAccrual, &totalWithdrawal)
	if err != nil {
		log.Print(err)
		return 0, 0, err
	}

	return
}
