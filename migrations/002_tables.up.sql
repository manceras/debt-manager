CREATE TYPE currency AS ENUM ('USD', 'EUR', 'GBP', 'JPY', 'CNY');

CREATE TABLE public.users (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	username text UNIQUE NOT NULL,
	email text UNIQUE NOT NULL,
	created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE public.lists (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	currency currency NOT NULL,
	title text NOT NULL,
	created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE public.users_lists (
	user_id uuid REFERENCES users(id) ON DELETE CASCADE,
	list_id uuid REFERENCES lists(id) ON DELETE CASCADE,
	PRIMARY KEY (user_id, list_id)
);

CREATE TABLE public.categories (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	name text NOT NULL,
	icon text,
	created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE public.payments (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	amount numeric(12, 2) NOT NULL,
	created_at TIMESTAMPTZ DEFAULT now(),
	photo_url text,
	payer_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
	list_id uuid REFERENCES lists(id) ON DELETE SET NULL
);

CREATE TABLE public.divisions (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	amount numeric(12, 2) NOT NULL,
	created_at TIMESTAMPTZ DEFAULT now(),
	owe_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
	payment_id uuid REFERENCES payments(id) ON DELETE CASCADE
);

CREATE TABLE public.payments_categories (
	payment_id uuid REFERENCES payments(id) ON DELETE CASCADE,
	category_id uuid REFERENCES categories(id) ON DELETE CASCADE,
	PRIMARY KEY (payment_id, category_id)
);

CREATE TABLE public.deposits (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	amount numeric(12, 2) NOT NULL,
	created_at TIMESTAMPTZ DEFAULT now(),
	payer_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
	payee_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
	list_id uuid REFERENCES lists(id) ON DELETE SET NULL
);

CREATE INDEX ON public.users_lists (user_id);
CREATE INDEX ON public.users_lists (list_id);
CREATE INDEX ON public.payments (list_id);
CREATE INDEX ON public.payments (payer_user_id);
CREATE INDEX ON public.divisions (payment_id);
CREATE INDEX ON public.divisions (owe_user_id);
CREATE INDEX ON public.deposits (list_id);
CREATE INDEX ON public.deposits (payer_user_id);
CREATE INDEX ON public.deposits (payee_user_id);
CREATE INDEX ON public.categories (name);
