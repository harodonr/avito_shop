-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    coins INTEGER DEFAULT 1000 NOT NULL
);

-- Создание таблицы товаров (мерча)
CREATE TABLE IF NOT EXISTS merchandise (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    price INTEGER NOT NULL
);

-- Добавление начальных данных в таблицу товаров
INSERT INTO merchandise (name, price) VALUES
    ('t-shirt', 80),
    ('cup', 20),
    ('book', 50),
    ('pen', 10),
    ('powerbank', 200),
    ('hoody', 300),
    ('umbrella', 200),
    ('socks', 10),
    ('wallet', 50),
    ('pink-hoody', 500)
ON CONFLICT (name) DO NOTHING; -- Не вставлять дубли, если они уже есть

-- Создание таблицы покупок (связь пользователя и товаров)
CREATE TABLE IF NOT EXISTS purchases (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    merchandise_id INTEGER REFERENCES merchandise(id) ON DELETE CASCADE,
    purchase_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы инвентаря с количеством
CREATE TABLE IF NOT EXISTS user_inventory (
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    merchandise_id INT REFERENCES merchandise(id) ON DELETE CASCADE,
    quantity INT DEFAULT 0,
    PRIMARY KEY (user_id, merchandise_id)
);

-- Создание таблицы переводов монет между пользователями
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    sender_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    receiver_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    amount INTEGER NOT NULL,
    transaction_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    -- Добавление уникального ограничения на комбинацию sender_id и receiver_id
);

-- Пример вставки данных в таблицу пользователей
INSERT INTO users (username, password_hash, coins) VALUES
    ('user1', 'password_hash_1', 1000),
    ('user2', 'password_hash_2', 1000),
    ('user3', 'password_hash_3', 1000)
ON CONFLICT (username) DO NOTHING; -- Не вставлять дубли

-- Пример записи транзакции (перевод монет между пользователями)
INSERT INTO transactions (sender_id, receiver_id, amount) VALUES
    (1, 2, 50),
    (2, 3, 100); -- Не вставлять дубли

-- Пример покупки товара
INSERT INTO purchases (user_id, merchandise_id) VALUES
    (1, 1),  -- user1 покупает t-shirt
    (1, 3),  -- user1 покупает book
    (2, 2);  -- user2 покупает cup

