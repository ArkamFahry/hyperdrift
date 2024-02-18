-- +goose Up
-- +goose StatementBegin

create or replace function storage.on_event_create()
    returns trigger as
$$
begin
    new.id = 'event' || '_' || storage.gen_random_ulid();
    new.version = 0;
    new.created_at = now();

    return new;
end;
$$ language plpgsql;

create table if not exists storage.events
(
    id             text          not null check ( storage.text_non_empty_trimmed_text(id) ),
    version        int default 0 not null check ( version >= 0 ),
    aggregate_type text          not null check ( storage.text_non_empty_trimmed_text(aggregate_type) ),
    aggregate_id   text          not null check ( storage.text_non_empty_trimmed_text(aggregate_id) ),
    event_type     text          not null check ( storage.text_non_empty_trimmed_text(event_type) ),
    payload        jsonb         null,
    created_at     timestamptz   not null,
    constraint events_id_primary_key primary key (id)
);

create or replace trigger event_on_create
    before insert
    on storage.events
    for each row
execute function storage.on_event_create();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop trigger if exists events_on_create on storage.events;

drop table if exists storage.events;

drop function if exists storage.on_event_create();

-- +goose StatementEnd
