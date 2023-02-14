CREATE TABLE IF NOT EXISTS current_balance
(
    user_ID INT,
    balance FLOAT DEFAULT 0,
    total_withdrawal FLOAT DEFAULT 0,
    FOREIGN KEY (user_ID) REFERENCES users (user_ID)
);