CREATE TABLE user
(
    id           bigint auto_increment primary key,
    username     varchar(32) not null,
    password     varchar(128) not null,
    created_at   datetime(3) default CURRENT_TIMESTAMP(3) not null,
    updated_at   datetime(3) default CURRENT_TIMESTAMP(3) not null
);