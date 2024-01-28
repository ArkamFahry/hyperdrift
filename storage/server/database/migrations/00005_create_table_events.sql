-- +goose Up
-- +goose StatementBegin

create table if not exists storage.events
(
    id text not null check ( storage.text_non_empty_trimmed_text(id) ),
    name text not null check ( storage.text_non_empty_trimmed_text(name) ),
    payload jsonb not null,
    status text        default 'pending'                  not null check (
        status in ('pending', 'in_progress', 'processing', 'completed', 'failed')
        ),
    producer text default 'hyperdrift-storage' not null check ( storage.text_non_empty_trimmed_text(producer) ),
    timestamp timestamptz default now() not null,
    constraint events_id_pk primary key (id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop table if exists storage.events;

-- +goose StatementEnd
