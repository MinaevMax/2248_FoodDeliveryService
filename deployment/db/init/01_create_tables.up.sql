CREATE EXTENSION IF NOT EXISTS pgcrypto;

DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS users CASCADE;

CREATE TABLE users (
    id          UUID        PRIMARY KEY     DEFAULT gen_random_uuid (),             --id пользователя
    login       TEXT        NOT NULL,                                               --логин пользователя
    password    VARCHAR(64) NOT NULL,                                               --пароль(sha256)
    is_active   BOOLEAN                     DEFAULT TRUE,                           --Флаг активного пользователя
    created_at TIMESTAMP    WITH    TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP   --Таймстемп регистрации
);

CREATE TABLE orders (
    id              UUID PRIMARY KEY    DEFAULT gen_random_uuid (),                 --id заказа
    user_id         UUID REFERENCES users (id),                                     --id пользователя
    status          VARCHAR(20)         DEFAULT 'UNDEFINED',                        --статус заказа UNDEFINED/PACKING/ARRIVING/COMPLETED/CANCELED
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,    --Таймстемп создания заказа
    completed_at    TIMESTAMP WITH TIME ZONE,                                       --Таймстемп успешного завершения заказа
    updated_at      TIMESTAMP WITH TIME ZONE                                        --Таймстемп последнего обновления информации
);
