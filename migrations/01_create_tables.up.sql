DROP TABLE IF EXISTS customers CASCADE;
DROP TABLE IF EXISTS couriers CASCADE;
DROP TABLE IF EXISTS orders CASCADE;

CREATE TABLE customers
(
    id            UUID PRIMARY KEY                  DEFAULT gen_random_uuid(), --id пользователя
    first_name    VARCHAR(32)              NOT NULL,                           --Имя
    last_name     VARCHAR(32)              NOT NULL,                           --Фамилия
    email         VARCHAR(64) UNIQUE       NOT NULL,                           --Email (уникальный)
    phone_number  VARCHAR(20) UNIQUE       NOT NULL,                           --Номер телефона (с +, уникальный)
    password_hash VARCHAR(64)              NOT NULL,                           --Пароль
    sex           VARCHAR(1),                                                  --Пол - M/F/null
    birthday      DATE,                                                        --Дата рождения
    is_active     BOOLEAN                           DEFAULT TRUE,              --Флаг активного пользователя
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, --Таймстемп регистрации
    updated_at    TIMESTAMP WITH TIME ZONE          DEFAULT CURRENT_TIMESTAMP  --Таймстемп последнего обновления информации
);

CREATE TABLE couriers
(
    id            UUID PRIMARY KEY                  DEFAULT gen_random_uuid(), --id пользователя
    first_name    VARCHAR(32)              NOT NULL,                           --Имя
    last_name     VARCHAR(32)              NOT NULL,                           --Фамилия
    email         VARCHAR(64) UNIQUE       NOT NULL,                           --Email (уникальный)
    phone_number  VARCHAR(20) UNIQUE       NOT NULL,                           --Номер телефона (с +, уникальный)
    password_hash VARCHAR(64)              NOT NULL,                           --Пароль
    sex           VARCHAR(1),                                                  --Пол - M/F/null
    birthday      DATE,                                                        --Дата рождения
    is_active     BOOLEAN                           DEFAULT TRUE,              --Флаг активного курьера
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, --Таймстемп регистрации
    updated_at    TIMESTAMP WITH TIME ZONE                                     --Таймстемп последнего обновления информации
);

CREATE TABLE orders
(
    id           UUID PRIMARY KEY                  DEFAULT gen_random_uuid(), --id заказа
    customer_id  UUID REFERENCES customers (id),                              --id пользователя
    courier_id   UUID REFERENCES couriers (id),                               --id курьера
    amount       DECIMAL(10, 2)           NOT NULL,                           --сумма заказа
    status       VARCHAR(20)                       DEFAULT 'UNDEFINED',       --статус заказа UNDEFINED/PACKING/ARRIVING/COMPLETED/CANCELED
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, --Таймстемп создания заказа
    completed_at TIMESTAMP WITH TIME ZONE,--Таймстемп успешного завершения заказа
    updated_at   TIMESTAMP WITH TIME ZONE--Таймстемп последнего обновления информации
);

--индексы