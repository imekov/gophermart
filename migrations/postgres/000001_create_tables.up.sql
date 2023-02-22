CREATE TABLE IF NOT EXISTS users
(
    user_ID INT GENERATED ALWAYS AS IDENTITY,
    login VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    PRIMARY KEY(user_ID)
);

CREATE TABLE IF NOT EXISTS statuses
(
    status_ID INT GENERATED ALWAYS AS IDENTITY,
    title VARCHAR(255) NOT NULL UNIQUE,
    PRIMARY KEY(status_ID)
);

INSERT INTO statuses (title) VALUES ('NEW'), ('PROCESSING'), ('INVALID'), ('PROCESSED')
ON CONFLICT (title) DO NOTHING;

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

CREATE TABLE IF NOT EXISTS orders
(
    user_ID INT,
    order_id VARCHAR(100) NOT NULL UNIQUE,
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

CREATE TABLE IF NOT EXISTS current_balance
(
    user_ID INT,
    balance FLOAT DEFAULT 0,
    total_withdrawal FLOAT DEFAULT 0,
    FOREIGN KEY (user_ID) REFERENCES users (user_ID)
);