CREATE TABLE IF NOT EXISTS statuses
(
    status_ID INT GENERATED ALWAYS AS IDENTITY,
    title VARCHAR(255) NOT NULL UNIQUE,
    PRIMARY KEY(status_ID)
);

INSERT INTO statuses (title) VALUES ('NEW'), ('PROCESSING'), ('INVALID'), ('PROCESSED')
                             ON CONFLICT (title) DO NOTHING