--up
create table users
(
    id         serial,
    email      text not null,
    name       text not null,
    pass_hash  bytea,
    created_at timestamp default now(),
    updated_at timestamp default now(),

    primary key (id)
);

--down
drop table users;
