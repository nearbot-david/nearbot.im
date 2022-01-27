-- auto-generated definition
create table balance
(
    id          serial
        constraint balance_pk
            primary key,
    telegram_id bigint              not null,
    amount      bigint    default 0 not null,
    created_at  timestamp default now(),
    updated_at  timestamp
);

alter table balance
    owner to moneybotuser;

create index balance_telegram_id_index
    on balance (telegram_id);

create unique index balance_telegram_id_uindex
    on balance (telegram_id);

