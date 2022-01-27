-- auto-generated definition
create table withdrawal
(
    id          serial
        constraint withdrawal_pk
            primary key,
    slug        varchar                    not null,
    telegram_id bigint                     not null,
    status      varchar                    not null,
    amount      bigint                     not null,
    address     varchar   default ''::character varying,
    created_at  timestamp default now()    not null,
    updated_at  timestamp default now()    not null,
    comment     text      default ''::text not null
);

alter table withdrawal
    owner to moneybotuser;

create index withdrawal_slug_index
    on withdrawal (slug);

create unique index withdrawal_slug_uindex
    on withdrawal (slug);

create index withdrawal_telegram_id_index
    on withdrawal (telegram_id);

