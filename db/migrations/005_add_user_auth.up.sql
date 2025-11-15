ALTER TABLE "users" 
  ADD COLUMN "email" varchar,
  ADD COLUMN "password_hash" varchar,
  ADD COLUMN "email_verified" boolean NOT NULL DEFAULT false;

-- Generate temporary emails for existing users
UPDATE "users" 
SET 
  "email" = 'user_' || id || '@migrated.local',
  "password_hash" = '$2a$10$TEMP_MIGRATION_HASH_REQUIRES_RESET'
WHERE "email" IS NULL;

-- Step 3: Now make email and password_hash NOT NULL
ALTER TABLE "users" 
  ALTER COLUMN "email" SET NOT NULL,
  ALTER COLUMN "password_hash" SET NOT NULL;

-- Step 4: Add unique constraint and index
CREATE UNIQUE INDEX idx_users_email ON "users" ("email");

