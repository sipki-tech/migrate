--up
create table sessions
(
    id         text not null,
    token      text not null,
    ip         inet not null,
    user_agent text not null,
    user_id    text not null,
    created_at timestamp default now(),
    updated_at timestamp default now(),

    unique (id),
    unique (token),
    primary key (id)
);

--down
drop table sessions;
