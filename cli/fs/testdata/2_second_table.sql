--up
create table second_table
(
    id       serial,
    username text not null,
    unique (username),
    primary key (id)
);

--down
drop table second_table;
