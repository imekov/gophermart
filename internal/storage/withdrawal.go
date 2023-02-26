package storage

import (
	"context"
	"log"

	"github.com/golang-module/carbon/v2"
)

type Withdrawals struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func (s PostgreConnect) InsertDataIntoWithdrawalsTable(order int, userID int, withdrawal float64, ctx context.Context) (error error) {
	tx, err := s.DBConnect.BeginTx(ctx, nil)
	if err != nil {
		log.Print(err)
		return err
	}
	defer tx.Rollback()

	sqlInsertData, err := tx.Prepare("INSERT INTO withdrawals (user_ID, order_ID, sum) VALUES ($1, $2, $3) ON CONFLICT (order_ID) DO NOTHING;")
	if err != nil {
		log.Print(err)
		return err
	}
	defer sqlInsertData.Close()

	_, err = sqlInsertData.Exec(userID, order, withdrawal)
	if err != nil {
		log.Print(err)
		return err
	}

	tx.Commit()

	return
}

func (s PostgreConnect) GetUserWithdrawals(ctx context.Context, userID int) (result []Withdrawals, err error) {

	result = []Withdrawals{}

	rows, err := s.DBConnect.QueryContext(ctx, "select order_ID, create_date, sum from withdrawals where user_id = $1;", userID)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var v Withdrawals

		err = rows.Scan(&v.Order, &v.ProcessedAt, &v.Sum)
		if err != nil {
			return []Withdrawals{}, err
		}

		v.ProcessedAt = carbon.Parse(v.ProcessedAt).ToRfc3339String()

		result = append(result, v)
	}

	err = rows.Err()
	if err != nil {
		return []Withdrawals{}, err
	}

	return
}
