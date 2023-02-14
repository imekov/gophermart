CREATE TABLE IF NOT EXISTS orders
(
    user_ID INT,
    order_ID BIGINT UNIQUE,
    status INT,
    create_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    accrual FLOAT,
    FOREIGN KEY (user_ID) REFERENCES users (user_ID),
    FOREIGN KEY (status) REFERENCES statuses (status_ID)
);

CREATE OR REPLACE FUNCTION update_total_accrual_balance()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS
$$
BEGIN
    IF NEW.accrual <> OLD.accrual THEN
        UPDATE current_balance
        SET balance = (SELECT SUM(orders.accrual) FROM orders WHERE orders.user_id = NEW.user_ID)
        WHERE user_ID = NEW.user_id;
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER total_accrual_balance
    AFTER UPDATE
    ON orders
    FOR EACH ROW
EXECUTE PROCEDURE update_total_accrual_balance();