--up
create table first_table
(
    id       serial,
    username text not null,
    unique (username),
    primary key (id)
);

--down
drop table first_table;
