CREATE VIEW group_balances AS (
    SELECT
    tx.group_id,
    COALESCE(gm.member_name, u.name) as user_name,
    SUM(
        CASE
        WHEN tx.by_user = s.split_user THEN 0
        ELSE -s.split_amount
        END
    ) +
    SUM(
        CASE
        WHEN tx.by_user = s.split_user THEN
            (
                SELECT 
                    SUM(s2.split_amount) 
                FROM splits s2 
                WHERE s2.transaction_id = tx.id 
                    AND s2.split_user != tx.by_user
            )
        ELSE 0
        END
    ) AS balance
    FROM splits s
    JOIN transactions tx ON s.transaction_id = tx.id
    JOIN group_members gm on gm.id = s.split_user
    JOIN users u on u.id = gm.user_id
    GROUP BY tx.group_id, user_name
    ORDER BY tx.group_id, user_name
);
