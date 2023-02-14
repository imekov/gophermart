package storage

import (
	"context"
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type PostgreConnect struct {
	DBConnect *sql.DB
}

type URLRow struct {
	UserID      string
	ShortURL    string
	OriginalURL string
}

type Orders struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual"`
	UploadedAt string  `json:"uploaded_at"`
}

type Withdrawals struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func GetNewConnection(db *sql.DB, dbConf string) *PostgreConnect {

	migration, err := migrate.New("file://migrations/postgres", dbConf)
	if err != nil {
		log.Print(err)
	}

	if err = migration.Up(); err != nil {
		log.Print(err)
	}

	return &PostgreConnect{DBConnect: db}
}

func (s PostgreConnect) CreateUser(login string, password string, ctx context.Context) (userID int, error error) {

	tx, err := s.DBConnect.BeginTx(ctx, nil)
	if err != nil {
		log.Print(err)
		return 0, err
	}
	defer tx.Rollback()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	err = tx.QueryRow("INSERT INTO users (login, password) VALUES ($1, $2) ON CONFLICT (login) DO NOTHING RETURNING user_ID;", login, hash).Scan(&userID)
	if err != nil {
		log.Print(err)
		return 0, err
	}

	sqlInsertData, err := tx.Prepare("INSERT INTO current_balance (user_ID) VALUES ($1);")
	if err != nil {
		log.Print(err)
		return 0, err
	}
	defer sqlInsertData.Close()

	_, err = sqlInsertData.Exec(userID)
	if err != nil {
		log.Print(err)
		return 0, err
	}

	tx.Commit()

	return userID, nil
}

func (s PostgreConnect) LoginUser(login string, password string, ctx context.Context) (userID int, error error) {

	tx, err := s.DBConnect.BeginTx(ctx, nil)
	if err != nil {
		log.Print(err)
		return 0, err
	}

	var hashPassword string

	err = tx.QueryRow("select user_ID, password from users where login = $1;", login).Scan(&userID, &hashPassword)
	if err != nil {
		log.Print(err)
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	if err != nil {
		return 0, errors.New("invalid username/password ")
	}

	return userID, nil
}

func (s PostgreConnect) IsUserExistByLogin(login string, ctx context.Context) (response bool) {
	response = false

	tx, err := s.DBConnect.BeginTx(ctx, nil)
	if err != nil {
		log.Print(err)
	}

	var countOfRows int

	err = tx.QueryRow("select COUNT(*) from users where login = $1;", login).Scan(&countOfRows)
	if err != nil {
		log.Print(err)
		return false
	}

	if countOfRows != 0 {
		response = true
	}

	return response
}

func (s PostgreConnect) IsUserExistByUserID(userID int, ctx context.Context) (response bool) {
	response = false

	tx, err := s.DBConnect.BeginTx(ctx, nil)
	if err != nil {
		log.Print(err)
	}

	var countOfRows int

	err = tx.QueryRow("select COUNT(*) from users where user_ID = $1;", userID).Scan(&countOfRows)
	if err != nil {
		log.Print(err)
		return false
	}

	if countOfRows != 0 {
		response = true
	}

	return response
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

		tm, err := time.Parse(time.RFC3339, v.UploadedAt)
		if err != nil {
			return []Orders{}, err
		}

		v.UploadedAt = tm.String()

		result = append(result, v)
	}

	err = rows.Err()
	if err != nil {
		return []Orders{}, err
	}

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

		tm, err := time.Parse(time.RFC3339, v.ProcessedAt)
		if err != nil {
			return []Withdrawals{}, err
		}

		v.ProcessedAt = tm.String()

		result = append(result, v)
	}

	err = rows.Err()
	if err != nil {
		return []Withdrawals{}, err
	}

	return
}
