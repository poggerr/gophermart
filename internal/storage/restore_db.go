package storage

var schema = `
CREATE TABLE IF NOT EXISTS main_user (
    id UUID PRIMARY KEY,
    username text UNIQUE NOT NULL,
    password text NOT NULL,
    balance double precision NOT NULL DEFAULT 0,
    withdrawn double precision NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY,
    order_number text UNIQUE NOT NULL,
    order_user UUID NOT NULL,
    uploaded_at timestamptz NOT NULL,
    accrual_service double precision,
    status text NOT NULL
);

CREATE TABLE IF NOT EXISTS withdrawals (
    id UUID PRIMARY KEY,
    order_number text UNIQUE NOT NULL,
    order_user UUID NOT NULL,
    sum double precision NOT NULL,
    processed_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_orders_user_uploaded_at ON orders(order_user, uploaded_at);
CREATE INDEX IF NOT EXISTS idx_withdrawals_user_processed_at ON withdrawals(order_user, processed_at);
`

func (strg *Storage) RestoreDB() {
	strg.DB.MustExec(schema)
}
