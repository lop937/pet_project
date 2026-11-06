-- Пользователи
CREATE TABLE IF NOT EXISTS users
(
    id     INTEGER PRIMARY KEY AUTOINCREMENT,
    fio    TEXT    NOT NULL DEFAULT '',
    rate   REAL    NOT NULL DEFAULT 1.0,
    active INTEGER NOT NULL DEFAULT 1
);

-- Продукты
CREATE TABLE IF NOT EXISTS products
(
    id     INTEGER PRIMARY KEY AUTOINCREMENT,
    tag    TEXT    NOT NULL DEFAULT '',
    name   TEXT    NOT NULL DEFAULT '',
    price  REAL NOT NULL DEFAULT 0,
    unit   TEXT    NOT NULL DEFAULT '',
    square REAL NOT NULL DEFAULT 0,
    attr   TEXT    NOT NULL DEFAULT ''
);

-- Проекты
CREATE TABLE IF NOT EXISTS projects
(
    id       INTEGER PRIMARY KEY AUTOINCREMENT,
    name     TEXT    NOT NULL DEFAULT '',
    date     TEXT    NOT NULL DEFAULT '',
    customer TEXT    NOT NULL DEFAULT '',
    status   TEXT    NOT NULL DEFAULT 'новый'
);

-- Системы (принадлежат проекту)
CREATE TABLE IF NOT EXISTS systems
(
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    name       TEXT    NOT NULL DEFAULT '',
    status     TEXT    NOT NULL DEFAULT 'новый',
    FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE
);

-- Работы
CREATE TABLE IF NOT EXISTS works
(
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    date         TEXT    NOT NULL,
    name         TEXT    NOT NULL DEFAULT '',
    price        REAL NOT NULL DEFAULT 0,
    user_id      INTEGER NOT NULL,
    product_id   INTEGER NOT NULL,
    project_id   INTEGER,
    system_id    INTEGER,
    service_type TEXT    NOT NULL DEFAULT 'монтаж',
    quantity     REAL NOT NULL DEFAULT 0,
    created_at   INTEGER          DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT,
    FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE RESTRICT,
    FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE SET NULL,
    FOREIGN KEY (system_id) REFERENCES systems (id) ON DELETE SET NULL
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_works_user ON works (user_id);
CREATE INDEX IF NOT EXISTS idx_works_date ON works (date);
CREATE INDEX IF NOT EXISTS idx_works_product ON works (product_id);
CREATE INDEX IF NOT EXISTS idx_works_project ON works (project_id);
CREATE INDEX IF NOT EXISTS idx_works_system ON works (system_id);
CREATE INDEX IF NOT EXISTS idx_projects_date ON projects (date);
CREATE INDEX IF NOT EXISTS idx_systems_project ON systems (project_id);