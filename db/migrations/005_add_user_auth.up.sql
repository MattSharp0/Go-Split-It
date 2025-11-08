ALTER TABLE "users" 
  ADD COLUMN "email" varchar UNIQUE NOT NULL,
  ADD COLUMN "password_hash" varchar NOT NULL,
  ADD COLUMN "email_verified" boolean NOT NULL DEFAULT false;

CREATE INDEX idx_users_email ON "users" ("email");

