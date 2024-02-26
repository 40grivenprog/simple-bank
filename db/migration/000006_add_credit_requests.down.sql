DROP INDEX IF EXISTS idx_status;
DROP INDEX IF EXISTS idx_unique_pending_status_per_user;
ALTER TABLE "credit_requests" DROP CONSTRAINT IF EXISTS credit_requests_username_fkey;
DROP TABLE IF EXISTS "credit_requests";
DROP TYPE IF EXISTS credit_requests_status;
