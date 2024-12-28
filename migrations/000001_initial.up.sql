CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE sessions (
	token TEXT PRIMARY KEY,
	data BYTEA NOT NULL,
	expiry TIMESTAMPTZ NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);

CREATE TABLE users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	username text NOT NULL,
	hashed_password text NOT NULL,
	pgp_key text DEFAULT NULL,
	prev_login TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE(username)
);

CREATE TABLE products (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	title TEXT NOT NULL,
	description TEXT NOT NULL,
	image_filename TEXT NOT NULL,
	vendor_id UUID REFERENCES users(id) NOT NULL,
	deleted_at TIMESTAMPTZ
);

CREATE TABLE delivery_methods (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	description TEXT NOT NULL,
	price int NOT NULL,
	product_id UUID REFERENCES products(id) NOT NULL,
	CHECK(price >= 0)
);

CREATE TABLE prices (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	quantity INT NOT NULL,
	price INT NOT NULL,
	product_id UUID REFERENCES products(id) NOT NULL,
	UNIQUE(quantity, product_id),
	CHECK(price >= 0)
);

CREATE TABLE orders (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	status TEXT NOT NULL,
	details TEXT NOT NULL,
	price_id UUID REFERENCES prices(id) NOT NULL,
	delivery_method_id UUID REFERENCES delivery_methods(id) NOT NULL,
	customer_id UUID REFERENCES users(id) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE invoices (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	address TEXT NOT NULL CHECK (LENGTH(address) = 95),
	order_id UUID REFERENCES orders(id) NOT NULL,
	xmr_price BIGINT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE(order_id)
);

CREATE TABLE reviews (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	grade INT NOT NULL,
	message TEXT NOT NULL,
	order_id UUID REFERENCES orders(id) NOT NULL,
	UNIQUE(order_id)
);

CREATE TABLE wallets (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	balance BIGINT NOT NULL DEFAULT 0,
	address TEXT NOT NULL CHECK (LENGTH(address) = 95),
	user_id UUID REFERENCES users(id) NOT NULL,
	UNIQUE(user_id)
);

CREATE TABLE withdrawals (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	amount BIGINT NOT NULL,
	dest_address TEXT NOT NULL CHECK (LENGTH(dest_address) = 95),
	status TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE transactions (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	hash TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE disputes (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	claim TEXT NOT NULL,
	order_id UUID REFERENCES orders(id) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE(order_id)
);

CREATE TABLE counter_disputes (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	claim TEXT NOT NULL,
	dispute_id UUID REFERENCES disputes(id) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE(dispute_id)
);

CREATE TYPE dispute_outcome AS ENUM ('vendor won', 'draw', 'customer won');

CREATE TABLE dispute_decisions (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	outcome dispute_outcome NOT NULL,
	reason TEXT NOT NULL,
	dispute_id UUID REFERENCES disputes(id) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE(dispute_id)
);

CREATE TABLE vendor_pledges (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	amount BIGINT NOT NULL,
	logo_filename TEXT NOT NULL,	
	user_id UUID REFERENCES users(id) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE(user_id)
);

CREATE TABLE decline_reasons (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	reason TEXT NOT NULL,
	order_id UUID REFERENCES orders(id) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE(order_id)
);

CREATE TABLE delivery_infos (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	info TEXT NOT NULL,
	order_id UUID REFERENCES orders(id) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE(order_id)
);

CREATE TABLE tickets (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	subject TEXT NOT NULL,
	message TEXT NOT NULL,
	author_id UUID REFERENCES users(id) NOT NULL,	
	is_open BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ticket_responses (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	message TEXT NOT NULL,
	ticket_id UUID REFERENCES tickets(id) NOT NULL,
	author_name TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE bans (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID REFERENCES users(id) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	UNIQUE(user_id)
);
