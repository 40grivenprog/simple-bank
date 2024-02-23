CREATE TYPE user_role AS ENUM ('base', 'admin');

ALTER TABLE "users" ADD COLUMN "role" user_role NOT NULL DEFAULT 'base';
