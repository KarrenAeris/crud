-- таблица товара
CREATE TABLE products
(
    id      BIGSERIAL PRIMARY KEY,
    name    TEXT      NOT NULL,
    price   INTEGER   NOT NULL CHECK (price > 0),
    qty     INTEGER   NOT NULL DEFAULT 0 CHECK (qty >= 0),
    active  BOOLEAN   NOT NULL DEFAULT TRUE,
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- таблица сотрудников
CREATE TABLE managers
(
    id        BIGSERIAL PRIMARY KEY,
    name      TEXT      NOT NULL,
    salary    INTEGER   NOT NULL DEFAULT 0,
    plan      INTEGER   NOT NULL DEFAULT 0,
    boss_id   BIGINT    REFERENCES managers,
    deparment TEXT,
    phone 	  TEXT      NOT NULL UNIQUE,
    password  TEXT,
    is_admin  BOOLEAN   NOT NULL DEFAULT TRUE,
    active    BOOLEAN   NOT NULL DEFAULT TRUE,
    created   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- таблица зарегистрированных покупателей
CREATE TABLE customers
(
    id        BIGSERIAL PRIMARY KEY,
    name      TEXT      NOT NULL,
    phone     TEXT      NOT NULL UNIQUE,
    password  TEXT      NOT NULL,
    active    BOOLEAN   NOT NULL DEFAULT TRUE,
    created   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- таблица продаж
CREATE TABLE sales
(
    id           BIGSERIAL PRIMARY KEY,
    manager_id   BIGINT    NOT NULL REFERENCES managers,
    customer_id  BIGINT    NOT NULL,
    created      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- конкретные позиции в порядке (чек)
CREATE TABLE sales_positions
(
    id          BIGSERIAL PRIMARY KEY,
    product_id BIGINT    NOT NULL REFERENCES products,
    sale_id     BIGINT    NOT NULL REFERENCES sales,
    price       INTEGER   NOT NULL CHECK ( price >= 0),
    qty         INTEGER   NOT NULL DEFAULT 0 CHECK (qty >= 0),
    created     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- таблица токенов зарегистрированных покупателей
CREATE TABLE customers_tokens (
    token       TEXT      NOT NULL UNIQUE,
    customer_id BIGINT    NOT NULL references customers,
    expire      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 hour',
    created     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- таблица токенов зарегистрированных продавцов
CREATE TABLE managers_tokens 
(
    token      TEXT      NOT NULL UNIQUE,
    manager_id BIGINT    NOT NULL references managers,
    expire     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 hour',
    created    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);