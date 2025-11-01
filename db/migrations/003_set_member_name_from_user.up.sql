-- Function to automatically set member_name from users.name when user_id is set
CREATE OR REPLACE FUNCTION set_group_member_name_from_user()
RETURNS TRIGGER AS $$
BEGIN
    -- If user_id is set but member_name is NULL, populate it from users table
    IF NEW.user_id IS NOT NULL AND NEW.member_name IS NULL THEN
        NEW.member_name = (
            SELECT name 
            FROM users 
            WHERE id = NEW.user_id
        );
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for INSERT operations
CREATE TRIGGER set_member_name_on_insert
BEFORE INSERT ON group_members
FOR EACH ROW
EXECUTE FUNCTION set_group_member_name_from_user();

-- Trigger for UPDATE operations (only when user_id is set and member_name is NULL)
CREATE TRIGGER set_member_name_on_update
BEFORE UPDATE ON group_members
FOR EACH ROW
WHEN (NEW.user_id IS NOT NULL AND NEW.member_name IS NULL)
EXECUTE FUNCTION set_group_member_name_from_user();

-- Backfill existing records: update member_name for existing group_members 
-- that have a user_id but NULL member_name
UPDATE group_members gm
SET member_name = u.name
FROM users u
WHERE gm.user_id = u.id
  AND gm.user_id IS NOT NULL
  AND gm.member_name IS NULL;