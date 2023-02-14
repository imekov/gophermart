CREATE TABLE IF NOT EXISTS withdrawals
(
    user_ID INT,
    order_id VARCHAR(100) NOT NULL UNIQUE,
    create_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sum FLOAT,
    FOREIGN KEY (user_ID) REFERENCES users (user_ID)
);


CREATE OR REPLACE FUNCTION update_total_withdrawals_balance()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS
$$
BEGIN
    UPDATE current_balance
    SET total_withdrawal = (SELECT SUM(withdrawals.sum) FROM withdrawals WHERE withdrawals.user_id = NEW.user_ID)
    WHERE user_ID = NEW.user_id;

    RETURN NEW;
END;
$$;

CREATE TRIGGER total_accrual_balance
    AFTER INSERT
    ON withdrawals
    FOR EACH ROW
EXECUTE PROCEDURE update_total_withdrawals_balance();