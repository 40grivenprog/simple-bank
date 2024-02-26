CREATE TYPE credit_requests_status AS ENUM ('approved', 'pending', 'cancelled');

CREATE TABLE "credit_requests" (
  "id" bigserial PRIMARY KEY,
  "status" credit_requests_status NOT NULL DEFAULT 'pending',
  "amount" int NOT NULL,
  "reason" varchar,
  "username" varchar NOT NULL,
  "currency" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

ALTER TABLE "credit_requests" ADD FOREIGN KEY ("username") REFERENCES "users" ("username");
CREATE UNIQUE INDEX idx_unique_pending_status_per_user ON credit_requests (username) WHERE status = 'pending';
CREATE INDEX idx_status ON credit_requests (status);
