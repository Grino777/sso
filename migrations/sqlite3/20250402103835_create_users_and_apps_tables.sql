-- +goose Up
-- +goose StatementBegin

CREATE TABLE
    roles (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name VARCHAR(50) UNIQUE NOT NULL
    );

INSERT INTO roles (name) VALUES ('user'), ('admin'), ('superadmin');

CREATE TABLE
    users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username VARCHAR(50) UNIQUE NOT NULL ,
        pass_hash VARCHAR(100) NOT NULL,
        role_id INTEGER NOT NULL DEFAULT 1,
        FOREIGN KEY (role_id) REFERENCES roles(id)
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
DROP TABLE roles;
DROP TABLE apps;

-- +goose StatementEnd