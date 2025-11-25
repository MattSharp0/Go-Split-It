CREATE TABLE "refresh_tokens" (
  "id" bigserial PRIMARY KEY,
  "token_hash" varchar NOT NULL,
  "user_id" bigint NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "revoked_at" timestamptz,
  "device_info" varchar,
  
  CONSTRAINT refresh_tokens_user_id_fkey FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_refresh_tokens_token_hash ON "refresh_tokens" ("token_hash");
CREATE INDEX idx_refresh_tokens_user_id ON "refresh_tokens" ("user_id");
CREATE INDEX idx_refresh_tokens_expires_at ON "refresh_tokens" ("expires_at");
CREATE INDEX idx_refresh_tokens_revoked_at ON "refresh_tokens" ("revoked_at");

