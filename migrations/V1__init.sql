-- V1__init.sql
-- Инициализация БД: расширения, enum-ы, таблицы, индексы, триггеры

-- Включаем генерирование UUID (pgcrypto)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ENUM types
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'product_status') THEN
    CREATE TYPE product_status AS ENUM ('ACTIVE','INACTIVE','ARCHIVED');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'order_status') THEN
    CREATE TYPE order_status AS ENUM ('CREATED','PAYMENT_PENDING','PAID','SHIPPED','COMPLETED','CANCELED');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'promo_discount_type') THEN
    CREATE TYPE promo_discount_type AS ENUM ('PERCENTAGE','FIXED_AMOUNT');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_operation_type') THEN
    CREATE TYPE user_operation_type AS ENUM ('CREATE_ORDER','UPDATE_ORDER');
  END IF;
END$$;


-- USERS (простая таблица для авторизации)
CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email varchar(255) UNIQUE NOT NULL,
  password_hash varchar(255) NOT NULL,
  role varchar(20) NOT NULL DEFAULT 'USER',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- PRODUCTS
CREATE TABLE IF NOT EXISTS products (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(255) NOT NULL,
  description varchar(4000),
  price numeric(12,2) NOT NULL CHECK (price > 0),
  stock integer NOT NULL CHECK (stock >= 0),
  category varchar(100) NOT NULL,
  status product_status NOT NULL DEFAULT 'ACTIVE',
  seller_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_products_status ON products (status);
CREATE INDEX IF NOT EXISTS idx_products_seller ON products (seller_id);

-- PROMO CODES
CREATE TABLE IF NOT EXISTS promo_codes (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code varchar(20) NOT NULL UNIQUE,
  discount_type promo_discount_type NOT NULL,
  discount_value numeric(12,2) NOT NULL CHECK (discount_value >= 0),
  min_order_amount numeric(12,2) NOT NULL DEFAULT 0,
  max_uses integer NOT NULL DEFAULT 0,
  current_uses integer NOT NULL DEFAULT 0,
  valid_from timestamptz,
  valid_until timestamptz,
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_promo_code_code ON promo_codes (code);

-- ORDERS
CREATE TABLE IF NOT EXISTS orders (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status order_status NOT NULL DEFAULT 'CREATED',
  promo_code_id uuid REFERENCES promo_codes(id),
  total_amount numeric(12,2) NOT NULL DEFAULT 0,
  discount_amount numeric(12,2) NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders (user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders (status);

-- ORDER ITEMS
CREATE TABLE IF NOT EXISTS order_items (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id uuid NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id uuid NOT NULL REFERENCES products(id),
  quantity integer NOT NULL CHECK (quantity > 0),
  price_at_order numeric(12,2) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items (order_id);

-- USER OPERATIONS (for rate limiting)
CREATE TABLE IF NOT EXISTS user_operations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  operation_type user_operation_type NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_operations_user_op ON user_operations (user_id, operation_type, created_at DESC);

-- REFRESH TOKENS (optional, для refresh-token management)
CREATE TABLE IF NOT EXISTS refresh_tokens (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash varchar(255) NOT NULL,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- TRIGGERS для updated_at
CREATE OR REPLACE FUNCTION set_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Attach trigger to tables that have updated_at
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE set_updated_at_column();

DROP TRIGGER IF EXISTS trg_products_updated_at ON products;
CREATE TRIGGER trg_products_updated_at BEFORE UPDATE ON products FOR EACH ROW EXECUTE PROCEDURE set_updated_at_column();

DROP TRIGGER IF EXISTS trg_promo_updated_at ON promo_codes;
CREATE TRIGGER trg_promo_updated_at BEFORE UPDATE ON promo_codes FOR EACH ROW EXECUTE PROCEDURE set_updated_at_column();

DROP TRIGGER IF EXISTS trg_orders_updated_at ON orders;
CREATE TRIGGER trg_orders_updated_at BEFORE UPDATE ON orders FOR EACH ROW EXECUTE PROCEDURE set_updated_at_column();

-- OPTIONAL: seed some data for demo (users: admin, seller, user)
INSERT INTO users (id, email, password_hash, role)
SELECT gen_random_uuid(), v.email, v.hash, v.role
FROM (VALUES
  ('admin@example.com', 'fake_hashed_password_admin', 'ADMIN'),
  ('seller@example.com','fake_hashed_password_seller','SELLER'),
  ('user@example.com','fake_hashed_password_user','USER')
) AS v(email, hash, role)
ON CONFLICT (email) DO NOTHING;