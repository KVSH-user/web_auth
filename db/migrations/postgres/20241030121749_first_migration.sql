-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
                       id SERIAL PRIMARY KEY,
                       email VARCHAR(255) UNIQUE NOT NULL,
                       password VARCHAR(255) NOT NULL,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                       is_active BOOLEAN DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS user_messages (
                               id SERIAL PRIMARY KEY,
                               user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                               message_text TEXT NOT NULL,
                               sender_type VARCHAR(10) CHECK (sender_type IN ('user', 'system')),
                               created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS anonymous_users (
                                 id SERIAL PRIMARY KEY,
                                 identifier UUID DEFAULT gen_random_uuid(),
                                 created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS anonymous_users;
DROP TABLE IF EXISTS user_messages;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
