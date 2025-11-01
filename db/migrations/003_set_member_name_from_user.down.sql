DROP TRIGGER IF EXISTS set_member_name_on_update ON group_members;
DROP TRIGGER IF EXISTS set_member_name_on_insert ON group_members;

DROP FUNCTION IF EXISTS set_group_member_name_from_user();