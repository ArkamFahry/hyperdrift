-- +goose Up
-- +goose StatementBegin

create table if not exists storage.buckets
(
    id                      text                      not null check ( storage.text_non_empty_trimmed_text(id) ),
    version                 int         default 0     not null check ( version >= 0 ),
    name                    text                      not null check ( storage.text_non_empty_trimmed_text(name) ),
    allowed_mime_types      text[]                    null check (
        storage.array_null_or_contains_empty_trimmed_text(allowed_mime_types)
            and
        storage.array_null_or_text_values_unique(allowed_mime_types)
        ),
    max_allowed_object_size bigint                    null check ( storage.bigint_null_or_non_zero_bigint(max_allowed_object_size) ),
    public                  boolean     default false not null,
    disabled                boolean     default false not null,
    locked                  boolean     default false not null,
    lock_reason             text check ( storage.text_null_or_non_empty_trimmed_text(lock_reason) ),
    locked_at               timestamptz               null,
    created_at              timestamptz default now() not null,
    updated_at              timestamptz               null,
    constraint buckets_id_primary_key primary key (id),
    constraint buckets_id_version_unique unique (id, version),
    constraint buckets_name_unique unique (name)
);

create index if not exists buckets_name_index on storage.buckets using btree (name);

create or replace trigger buckets_on_create
    before insert
    on storage.buckets
    for each row
execute function storage.on_create();

create or replace trigger buckets_on_update
    before update
    on storage.buckets
    for each row
execute function storage.on_update();


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop trigger if exists buckets_increment_version on storage.buckets;

drop trigger if exists buckets_set_updated_at on storage.buckets;

drop index if exists storage.buckets_name_idx;

drop table if exists storage.buckets;

-- +goose StatementEnd