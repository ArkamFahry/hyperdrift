-- +goose Up
-- +goose StatementBegin

create or replace function storage.objects_search(bucket_name text, path_prefix text,
                                                  levels int default 1, limits int default 100, offsets int default 0)
    returns table
            (
                id               text,
                version          int,
                name             text,
                bucket_id        text,
                bucket_name      text,
                content_type     text,
                size             bigint,
                public           boolean,
                metadata         jsonb,
                upload_status    text,
                last_accessed_at timestamptz,
                created_at       timestamptz,
                updated_at       timestamptz
            )
    language plpgsql
as
$$
begin
    return query
        with files_folders as (select path_tokens[levels] as folder
                               from storage.objects
                               where objects.name ilike path_prefix || '%'
                                 and objects.bucket_id = (select id from storage.buckets where name = bucket_name)
                               group by folder
                               limit limits offset offsets)
        select objects.id               as id,
               objects.version          as version,
               files_folders.folder     as name,
               objects.bucket_id        as bucket_id,
               bucket_name              as bucket_name,
               objects.content_type     as content_type,
               objects.size             as size,
               objects.public           as public,
               objects.metadata         as metadata,
               objects.upload_status    as upload_status,
               objects.last_accessed_at as last_accessed_at,
               objects.created_at       as created_at,
               objects.updated_at       as updated_at
        from files_folders
                 left join storage.objects
                           on path_prefix || files_folders.folder = objects.name and
                              objects.bucket_id = (select id from storage.buckets where name = bucket_name);
end
$$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop function if exists storage.objects_search(text, text, int, int, int);

-- +goose StatementEnd
