create table products
(
    Code        bigint                                               not null,
    StoreID     int                                                  not null,
    GTIN        bigint unsigned              default '0'             not null,
    Name        varchar(250) charset utf8mb4 default '!БЕЗ НАЗВАНИЯ' not null,
    Count       smallint unsigned            default '0'             not null,
    Price       int unsigned                 default '0'             not null,
    Producer    varchar(250) charset utf8mb4                         not null,
    Country     varchar(250) charset utf8mb4                         not null,
    Description text charset utf8mb4                                 not null,
    primary key (Code, StoreID),
    constraint Code_StoreID_UNIQUE
        unique (Code, StoreID)
)
    collate = utf8mb4_unicode_ci;

create fulltext index ft_name_description
    on products (Name, Description);

create table promotions
(
    product_code bigint               not null
        primary key,
    product_name varchar(250)      not null,
    discount     int               not null,
    is_percent   tinyint default 0 not null
);

create table stores
(
    id             int auto_increment
        primary key,
    address        varchar(255)                           not null,
    upload_time    datetime default '1000-01-01 00:00:00' not null,
    pos_lat        double                                 not null,
    pos_lon        double                                 not null,
    mobile         varchar(45)                            not null,
    email          varchar(45)                            not null,
    booking_enable tinyint                                not null,
    schedule       varchar(100) charset utf32             not null
);

create table booking
(
    id         int auto_increment
        primary key,
    store_id   int                                   not null,
    created_at datetime    default CURRENT_TIMESTAMP not null,
    status     varchar(45) default 'create'          not null,
    username   varchar(45) charset utf32             not null,
    phone      varchar(45)                           not null,
    message    text charset utf32                    not null,
    constraint book_to_store
        foreign key (store_id) references stores (id)
            on delete cascade
);

create index book_to_store_idx
    on booking (store_id);

create table booking_products
(
    id         int auto_increment
        primary key,
    booking_id int           not null,
    code_stu   bigint           not null,
    name       varchar(250)  not null,
    quantity   int default 1 not null,
    price      int           not null,
    constraint booking_prod_booking_id_key
        foreign key (booking_id) references booking (id)
            on delete cascade
);

create index booking_prod_book_id_key_idx
    on booking_products (booking_id);

