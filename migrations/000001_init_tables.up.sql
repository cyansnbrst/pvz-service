CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role VARCHAR(50) CHECK (role IN ('employee', 'moderator')) NOT NULL
);

CREATE TABLE pvzs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    city VARCHAR(50) CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань')) NOT NULL,
    registration_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE receptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    pvz_id UUID REFERENCES pvzs(id) NOT NULL,
    status VARCHAR(50) CHECK (status IN ('in_progress', 'close')) DEFAULT 'in_progress'
);

CREATE UNIQUE INDEX unique_in_progress_pvz ON receptions (pvz_id, status) WHERE status = 'in_progress';

CREATE INDEX idx_receptions_pvz_id ON receptions (pvz_id);
CREATE INDEX idx_receptions_date_time ON receptions (date_time);

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    type VARCHAR(50) NOT NULL CHECK (type IN ('электроника', 'одежда', 'обувь')),
    reception_id UUID REFERENCES receptions(id) NOT NULL
);

CREATE INDEX idx_products_reception_id ON products (reception_id);

