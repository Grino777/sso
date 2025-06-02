-- +goose Up
-- +goose StatementBegin
CREATE TABLE users_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    app_id INTEGER NOT NULL,
    user_ip VARCHAR(50) DEFAULT "unknown",
    loggined_at VARCHAR(50) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users_logins;
-- +goose StatementEnd
