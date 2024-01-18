-- +goose Up
-- +goose StatementBegin

create schema if not exists storage;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop schema if exists storage cascade;

-- +goose StatementEnd
