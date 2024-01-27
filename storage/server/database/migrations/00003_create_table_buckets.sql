-- +goose Up
-- +goose StatementBegin

create table if not exists storage.buckets
(
    id                      text                      not null check ( storage.text_non_empty_trimmed_text(id) ),
    name                    text                      not null check ( storage.text_non_empty_trimmed_text(name) ),
    allowed_content_types   text[]                    null check (
        storage.array_null_or_contains_empty_trimmed_text(allowed_content_types)
            and
        storage.array_null_or_text_values_unique(allowed_content_types)
        ),
    max_allowed_object_size bigint                    null check ( storage.bigint_null_or_non_zero_bigint(max_allowed_object_size) ),
    public                  boolean     default false not null,
    disabled                boolean     default false not null,
    locked                  boolean     default false not null,
    lock_reason             text check ( storage.text_null_or_non_empty_trimmed_text(lock_reason) ),
    locked_at               timestamptz               null,
    created_at              timestamptz default now() not null,
    updated_at              timestamptz               null,
    constraint buckets_id_pk primary key (id),
    constraint buckets_name_uq unique (name)
);

create index if not exists buckets_name_idx on storage.buckets using btree (name);

create or replace trigger buckets_set_updated_at
    before update
    on storage.buckets
    for each row
execute function storage.set_updated_at();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop trigger if exists buckets_set_updated_at on storage.buckets;

drop index if exists storage.buckets_name_idx;

drop table if exists storage.buckets;

-- +goose StatementEnd