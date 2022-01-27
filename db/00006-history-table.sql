-- auto-generated definition
create table history
(
    id         serial
        constraint history_pk
            primary key,
    item_type  varchar                                 not null,
    model_id   bigint                                  not null,
    slug       varchar                                 not null,
    "from"     varchar   default ''::character varying not null,
    "to"       varchar   default ''::character varying not null,
    amount     bigint                                  not null,
    status     varchar                                 not null,
    cause      varchar   default ''::character varying not null,
    created_at timestamp default now()                 not null
);

alter table history
    owner to moneybotuser;

create index history_from_index
    on history ("from");

create index history_item_type_slug_index
    on history (item_type, slug);

create index history_to_index
    on history ("to");

