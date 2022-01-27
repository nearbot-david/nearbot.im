-- auto-generated definition
create table transfer
(
    id         serial
        constraint transfer_pk
            primary key,
    slug       varchar                 not null,
    "from"     bigint                  not null,
    "to"       bigint    default 0,
    amount     bigint                  not null,
    status     varchar                 not null,
    created_at timestamp default now() not null,
    updated_at timestamp default now() not null,
    message_id varchar
);

alter table transfer
    owner to moneybotuser;

create unique index transfer_slug_uindex
    on transfer (slug);

create index transfer_slug_index
    on transfer (slug);

