-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS geodata;

CREATE TABLE IF NOT EXISTS
    geodata.layers (
        id uuid default gen_random_uuid() not null primary key,
        name text not null,
        description text,
        "table" text not null unique,
        crs int not null,
        attribution text
    );
-- +goose StatementEnd
