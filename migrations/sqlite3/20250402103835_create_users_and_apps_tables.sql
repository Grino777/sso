-- +goose Up
-- +goose StatementBegin

CREATE TABLE
    users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username VARCHAR(50) UNIQUE NOT NULL ,
        pass_hash VARCHAR(100) NOT NULL
    );

CREATE TABLE
    apps (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name VARCHAR(50) NOT NULL,
        secret TEXT NOT NULL
    );

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE users;

DROP TABLE apps;

-- +goose StatementEnd