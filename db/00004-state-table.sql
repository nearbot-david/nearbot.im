-- auto-generated definition
create table state
(
    id          serial
        constraint state_pk
            primary key,
    telegram_id bigint                  not null,
    state       varchar                 not null,
    updated_at  timestamp default now() not null,
    message_id  bigint    default 0     not null
);

alter table state
    owner to moneybotuser;

