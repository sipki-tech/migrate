--up
create table third_table
(
    id       serial,
    username text not null,
    unique (username),
    primary key (id)
);

--down
drop table third_table;
