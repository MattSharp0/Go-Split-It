-- View to aggregate net balances per user across all groups
CREATE VIEW user_balances_net AS (
    SELECT
        gm.user_id,
        SUM(gbn.net_balance)::numeric(10, 2) AS net_balance
    FROM group_balances_net gbn
    JOIN group_members gm ON gm.id = gbn.user_id
    WHERE gm.user_id IS NOT NULL
    GROUP BY gm.user_id
    ORDER BY gm.user_id
);

-- View to aggregate balances between a user and other members across all groups
-- This shows the net balance between a user and each member (aggregated across all groups)
-- Positive values mean the member owes the user, negative means the user owes the member
CREATE VIEW user_balances_by_member AS (
    WITH creditor_balances AS (
        SELECT
            creditor_gm.user_id AS user_id,
            debtor_gm.user_id AS member_user_id,
            SUM(gb.total_owed)::numeric(10, 2) AS net_amount
        FROM group_balances gb
        JOIN group_members creditor_gm ON creditor_gm.id = gb.creditor
        JOIN group_members debtor_gm ON debtor_gm.id = gb.debtor
        WHERE creditor_gm.user_id IS NOT NULL
          AND debtor_gm.user_id IS NOT NULL
        GROUP BY creditor_gm.user_id, debtor_gm.user_id
    ),
    debtor_balances AS (
        SELECT
            debtor_gm.user_id AS user_id,
            creditor_gm.user_id AS member_user_id,
            -SUM(gb.total_owed)::numeric(10, 2) AS net_amount
        FROM group_balances gb
        JOIN group_members creditor_gm ON creditor_gm.id = gb.creditor
        JOIN group_members debtor_gm ON debtor_gm.id = gb.debtor
        WHERE creditor_gm.user_id IS NOT NULL
          AND debtor_gm.user_id IS NOT NULL
        GROUP BY debtor_gm.user_id, creditor_gm.user_id
    )
    SELECT
        user_id,
        member_user_id,
        SUM(net_amount)::numeric(10, 2) AS net_balance
    FROM (
        SELECT * FROM creditor_balances
        UNION ALL
        SELECT * FROM debtor_balances
    ) combined
    GROUP BY user_id, member_user_id
    ORDER BY user_id, member_user_id
);

