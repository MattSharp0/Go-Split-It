DROP INDEX IF EXISTS idx_users_email;
ALTER TABLE "users" 
  DROP COLUMN IF EXISTS "email_verified",
  DROP COLUMN IF EXISTS "password_hash",
  DROP COLUMN IF EXISTS "email";

