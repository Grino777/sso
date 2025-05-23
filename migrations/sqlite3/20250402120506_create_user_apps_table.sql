-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    user_apps (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        app_id INTEGER NOT NULL,
        is_blocked INTEGER NOT NULL DEFAULT 0 CHECK (is_blocked in (0, 1)),
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
        FOREIGN KEY (app_id) REFERENCES apps (id) ON DELETE CASCADE
    );

CREATE TABLE
    user_token (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        user_token VARCHAR(100) NOT NULL,
        expired_at TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES user (id) ON DELETE CASCADE
    );

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE user_apps;
DROP TABLE user_token;
-- +goose StatementEnd