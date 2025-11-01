CREATE VIEW group_balances AS (
    SELECT
        tx.group_id,
        tx.by_user as creditor,
        s.split_user as debtor,
        SUM(
            CASE WHEN tx.by_user = s.split_user THEN 0
            ELSE s.split_amount END
        )::numeric(10,2) as total_owed -- Sum returns unconstrained numeric
    FROM splits s
    JOIN transactions tx on tx.id = s.transaction_id
    WHERE tx.by_user != s.split_user
    GROUP BY tx.group_id, tx.by_user, s.split_user
    ORDER BY tx.group_id, tx.by_user, s.split_user
);

CREATE VIEW group_balances_net AS (
	WITH transaction_credits AS (
        SELECT
            tx.group_id,
            tx.by_user as user_id,
            SUM(COALESCE(s.split_amount, 0)) as net_amount
        FROM transactions tx
        LEFT JOIN splits s ON s.transaction_id = tx.id AND s.split_user != tx.by_user
        GROUP BY tx.group_id, tx.by_user
    ),
    split_debits AS (
        SELECT
            tx.group_id,
            s.split_user as user_id,
            -SUM(s.split_amount) as net_amount
        FROM splits s
        JOIN transactions tx ON s.transaction_id = tx.id
        WHERE tx.by_user != s.split_user
        GROUP BY tx.group_id, s.split_user
    )
    SELECT
        group_id,
        user_id,
        SUM(net_amount)::numeric(10, 2) AS net_balance -- Sum returns unconstrained numeric
    FROM (
        SELECT * FROM transaction_credits
        UNION ALL
        SELECT * FROM split_debits
    ) combined
    GROUP BY group_id, user_id
    ORDER BY group_id, user_id
);

CREATE INDEX transactions_group_byuser_idx ON "transactions" (group_id, by_user);
CREATE INDEX splits_transaction_splituser_idx ON "splits" (transaction_id, split_user);