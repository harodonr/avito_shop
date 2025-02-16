-- Создание таблицы сотрудников
CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы товаров (мерча)
CREATE TABLE merchandise (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    price INTEGER NOT NULL
);

-- Добавление товаров
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
('pink-hoody', 500);

-- Создание таблицы кошельков сотрудников
CREATE TABLE wallets (
    employee_id INTEGER REFERENCES employees(id) ON DELETE CASCADE,
    coins INTEGER DEFAULT 1000,
    PRIMARY KEY (employee_id)
);

-- Создание таблицы транзакций (перевод монет между сотрудниками)
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    sender_id INTEGER REFERENCES employees(id) ON DELETE CASCADE,
    receiver_id INTEGER REFERENCES employees(id) ON DELETE CASCADE,
    amount INTEGER NOT NULL,
    transaction_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы покупок
CREATE TABLE purchases (
    employee_id INTEGER REFERENCES employees(id) ON DELETE CASCADE,
    merchandise_id INTEGER REFERENCES merchandise(id) ON DELETE CASCADE,
    purchase_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (employee_id, merchandise_id)
);
