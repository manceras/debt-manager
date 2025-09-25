-- Drop types
DROP TYPE IF EXISTS currency;

-- Drop schema
DROP SCHEMA IF EXISTS app CASCADE;

-- Drop roles (optional, be careful if shared)
DROP ROLE IF EXISTS app_auth;
DROP ROLE IF EXISTS app_admin;
