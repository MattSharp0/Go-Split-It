DROP TRIGGER IF EXISTS set_modified_at_users on "users";
DROP TRIGGER IF EXISTS set_modified_at_transactions on "transactions";
DROP TRIGGER IF EXISTS set_modified_at_splits on "splits";
DROP TRIGGER IF EXISTS set_member_name_on_user_delete on "group_members";

DROP FUNCTION IF EXISTS update_modified_at();
DROP FUNCTION IF EXISTS set_group_member_name_on_user_delete();

DROP TABLE IF EXISTS "splits";
DROP TABLE IF EXISTS "transactions";
DROP TABLE IF EXISTS "group_members";
DROP TABLE IF EXISTS "groups";
DROP TABLE IF EXISTS "users";