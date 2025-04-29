CREATE TABLE "users" (
  "id" bigserial PRIMARY KEY,
  "name" varchar NOT NULL,
  -- "email" varchar NOT NULL,
  -- "password" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "modified_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "group_members" (
  "id" bigserial PRIMARY KEY,
  "group_id" bigint NOT NULL,
  "member_name" varchar,
  "user_id" bigint,
  "created_at" timestamptz NOT NULL DEFAULT (now()),

  CONSTRAINT group_members_user_or_name_not_null CHECK ("user_id" IS NOT NULL OR "member_name" IS NOT NULL)
);

CREATE TABLE "groups" (
  "id" bigserial PRIMARY KEY,
  "name" varchar NOT NULL
);

CREATE TABLE "transactions" (
  "id" bigserial PRIMARY KEY,
  "group_id" bigint NOT NULL,
  "name" varchar NOT NULL,
  "transaction_date" date NOT NULL DEFAULT (CURRENT_DATE),
  "amount" numeric(10,2) NOT NULL,
  "category" varchar,
  "note" varchar,
  "by_user" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "modified_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "splits" (
  "id" bigserial PRIMARY KEY,
  "transaction_id" bigint NOT NULL,
  "tx_amount" numeric(10,2) NOT NULL,
  "split_percent" decimal(5,4) NOT NULL,
  "split_amount" numeric(10,2) NOT NULL,
  "split_user" bigint, -- should only be null when group member is deleted to maintain split total
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "modified_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX ON "users" ("name");

CREATE INDEX ON "group_members" ("group_id");

CREATE INDEX ON "group_members" ("user_id");

CREATE UNIQUE INDEX group_members_user_id_unique ON "group_members" ("group_id", "user_id") WHERE "user_id" IS NOT NULL;

CREATE UNIQUE INDEX group_members_member_name_unique ON "group_members" ("group_id", "member_name") WHERE "member_name" IS NOT NULL;

CREATE INDEX ON "transactions" ("transaction_date");

CREATE INDEX ON "transactions" ("group_id");

CREATE INDEX ON "splits" ("transaction_id");

CREATE INDEX ON "splits" ("split_user");

ALTER TABLE "group_members" ADD FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON DELETE CASCADE; -- Group member is deleted if group is deleted

ALTER TABLE "group_members" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE SET NULL; -- Group member is set to null if user is deleted

ALTER TABLE "transactions" ADD FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON DELETE CASCADE; -- Transaction is deleted if group is deleted

ALTER TABLE "transactions" ADD FOREIGN KEY ("by_user") REFERENCES "group_members" ("id") ON DELETE CASCADE; -- Transaction is deleted if group member is deleted

ALTER TABLE "splits" ADD FOREIGN KEY ("transaction_id") REFERENCES "transactions" ("id") ON DELETE CASCADE; -- Split is deleted if transaction is deleted

ALTER TABLE "splits" ADD FOREIGN KEY ("split_user") REFERENCES "group_members" ("id") ON DELETE SET NULL; -- Split user is set to null if group member is deleted to maintain transaction integrity. Should set to a deleted user?

ALTER TABLE "splits" ADD CONSTRAINT "split_percent_valid_range" CHECK (split_percent >= 0 AND split_percent <= 1.0); -- Check that split_percent is between 0 and 1 (inlcusive)

CREATE OR REPLACE FUNCTION update_modified_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_modified_at_users
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_modified_at();

CREATE TRIGGER set_modified_at_transactions
BEFORE UPDATE ON transactions
FOR EACH ROW
EXECUTE FUNCTION update_modified_at();

CREATE TRIGGER set_modified_at_splits
BEFORE UPDATE ON splits
FOR EACH ROW
EXECUTE FUNCTION update_modified_at();

CREATE OR REPLACE FUNCTION set_group_member_name_on_user_delete()
RETURNS TRIGGER AS $$
BEGIN
    -- If member_name is NULL, set it to "user_name + user_id"
    IF NEW.member_name IS NULL THEN
        NEW.member_name = (
            SELECT CONCAT(name, ' (ID: ', OLD.user_id, ')')
            FROM users
            WHERE id = OLD.user_id
        );
    END IF;

    -- Return the updated row
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_member_name_on_user_delete
BEFORE UPDATE ON group_members
FOR EACH ROW
WHEN (OLD.user_id IS NOT NULL AND NEW.user_id IS NULL)
EXECUTE FUNCTION set_group_member_name_on_user_delete();