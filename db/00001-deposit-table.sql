-- auto-generated definition
create table deposit
(
    id             serial
        constraint deposit_pk
            primary key,
    slug           varchar                                    not null,
    telegram_id    bigint                                     not null,
    payment_method varchar,
    amount         bigint                                     not null,
    status         varchar   default 'NEW'::character varying not null,
    created_at     timestamp default now()                    not null,
    updated_at     timestamp default now()                    not null,
    message_id     bigint    default 0
);

alter table deposit
    owner to moneybotuser;

create unique index deposit_slug_uindex
    on deposit (slug);

create index deposit_slug_index
    on deposit (slug);

