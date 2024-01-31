-- +goose Up
-- +goose StatementBegin

create or replace function storage.text_non_empty_trimmed_text(val text) returns boolean as
$$
begin
    return trim(val) <> '';
end;
$$ language plpgsql;

create or replace function storage.text_null_or_non_empty_trimmed_text(val text) returns boolean as
$$
begin
    if val is null then
        return true;
    end if;

    return trim(val) <> '';
end;
$$ language plpgsql;

create or replace function storage.array_null_or_contains_empty_trimmed_text(val text[]) returns boolean as
$$
declare
    i int;
begin
    if val is null then
        return true;
    end if;

    for i in array_lower(val, 1) .. coalesce(array_upper(val, 1), 0)
        loop
            if trim(val[i]) = '' then
                return false;
            end if;
        end loop;

    return true;
end;
$$ language plpgsql;

create or replace function storage.array_null_or_text_values_unique(val text[])
    returns boolean as
$$
begin
    if val is null then
        return true;
    end if;

    return array_length(val, 1) = array_length(array(select distinct unnest(val)), 1);
end;
$$ language plpgsql;

create or replace function storage.bigint_null_or_non_zero_bigint(val bigint)
    returns boolean as
$$
begin
    if val is null then
        return true;
    end if;

    return val <> 0;
end
$$ language plpgsql;



create or replace function storage.set_updated_at()
    returns trigger as
$$
begin
    if new is distinct from old then
        new.updated_at = now();
    end if;

    return new;
end;
$$ language plpgsql;

create or replace function storage.increment_version()
    returns trigger as
$$
begin
    if new is distinct from old then
        new.version = new.version + 1;
    end if;

    return new;
end;
$$ language plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop function if exists storage.increment_version();

drop function if exists storage.set_updated_at();

drop function if exists storage.bigint_null_or_non_zero_bigint();

drop function if exists storage.array_null_or_text_values_unique();

drop function if exists storage.array_null_or_contains_empty_trimmed_text();

drop function if exists storage.text_null_or_non_empty_trimmed_text();

drop function if exists storage.text_non_empty_trimmed_text();

-- +goose StatementEnd
