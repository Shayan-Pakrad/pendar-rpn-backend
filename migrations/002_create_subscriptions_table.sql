CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE REFERENCES users(username),
    remaining_requests INT NOT NULL,
    expiry_date DATE NOT NULL
);