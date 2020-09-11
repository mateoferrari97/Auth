CREATE TABLE IF NOT EXISTS user
(
    id           bigint auto_increment primary key,
    _id          varchar(128) not null,
    firstname    varchar(128) not null,
    lastname     varchar(128) not null,
    created_at   datetime(3) default CURRENT_TIMESTAMP(3) not null,
    updated_at   datetime(3) default CURRENT_TIMESTAMP(3) not null
);

CREATE TABLE IF NOT EXISTS login
(
    id           bigint auto_increment primary key,
    email        varchar(128) not null,
    password     varchar(128) not null,
    user_id      bigint  not null,
    constraint login_user_id_fk
        foreign key (user_id) references user (id)
);