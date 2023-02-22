package storage

import (
	"context"
	"log"

	"github.com/golang-module/carbon/v2"
)

type Orders struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual"`
	UploadedAt string  `json:"uploaded_at"`
}

func (s PostgreConnect) GetUserOrders(ctx context.Context, userID int) (result []Orders, err error) {

	result = []Orders{}

	rows, err := s.DBConnect.QueryContext(ctx, "select order_ID, (SELECT title FROM statuses where statuses.status_id= orders.status), create_date, accrual from orders where user_id = $1;", userID)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var v Orders

		err = rows.Scan(&v.Number, &v.Status, &v.UploadedAt, &v.Accrual)
		if err != nil {
			return []Orders{}, err
		}

		v.UploadedAt = carbon.Parse(v.UploadedAt).ToRfc3339String()

		result = append(result, v)
	}

	err = rows.Err()
	if err != nil {
		return []Orders{}, err
	}

	return
}

func (s PostgreConnect) IsOrderExistByOrderID(order int, ctx context.Context) (userID int) {

	tx, err := s.DBConnect.BeginTx(ctx, nil)
	if err != nil {
		log.Print(err)
		return 0
	}

	err = tx.QueryRow("select user_ID from orders where order_ID = $1;", order).Scan(&userID)
	if err != nil {
		return 0
	}

	return userID
}

func (s PostgreConnect) InsertDataIntoOrdersTable(order int, userID int, status string, accrual float32, ctx context.Context) (error error) {

	tx, err := s.DBConnect.BeginTx(ctx, nil)
	if err != nil {
		log.Print(err)
		return err
	}
	defer tx.Rollback()

	sqlInsertData, err := tx.Prepare("INSERT INTO orders (user_ID, order_ID, status, accrual) VALUES ((SELECT user_ID from users WHERE user_ID=$1), $2, (SELECT status_ID from statuses WHERE title=$3), $4) ON CONFLICT (order_ID) DO NOTHING;")
	if err != nil {
		log.Print(err)
		return err
	}
	defer sqlInsertData.Close()

	_, err = sqlInsertData.Exec(userID, order, status, accrual)
	if err != nil {
		log.Print(err)
		return err
	}

	tx.Commit()

	return
}

func (s PostgreConnect) GetAllNewOrders(ctx context.Context) (result []string, err error) {

	rows, err := s.DBConnect.QueryContext(ctx, "select order_ID from orders where status = (select status_id from statuses where title = 'NEW') or status = (select status_id from statuses where title = 'PROCESSING');")
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var v string
		err = rows.Scan(&v)
		if err != nil {
			return []string{}, err
		}

		result = append(result, v)
	}

	err = rows.Err()
	if err != nil {
		return []string{}, err
	}

	return
}

func (s PostgreConnect) UpdateOrderInformation(ctx context.Context, orderNum string, status string, accrual float64) (err error) {

	tx, err := s.DBConnect.BeginTx(ctx, nil)
	if err != nil {
		log.Print(err)
		return err
	}
	defer tx.Rollback()

	sqlInsertData, err := tx.Prepare("UPDATE orders SET status = (select status_id from statuses where title = $1), accrual = $2 WHERE order_ID = $3;")
	if err != nil {
		log.Print(err)
		return err
	}
	defer sqlInsertData.Close()

	_, err = sqlInsertData.Exec(status, accrual, orderNum)
	if err != nil {
		log.Print(err)
		return err
	}

	tx.Commit()

	return
}
