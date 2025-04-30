CREATE VIEW group_balances AS (
    SELECT
        tx.group_id,
        tx.by_user as creditor,
        s.split_user as debtor,
        SUM(
            CASE WHEN tx.by_user = s.split_user THEN 0
            ELSE s.split_amount END
        ) as total_owed
    FROM splits s
    JOIN transactions tx on tx.id = s.transaction_id
    WHERE tx.by_user != s.split_user
    GROUP BY tx.group_id, tx.by_user, s.split_user
    ORDER BY tx.group_id, tx.by_user, s.split_user
);

CREATE VIEW group_balances_simple AS (
    WITH gbc as (
        SELECT
            gb.group_id,
            gb.creditor,
            gb.debtor,
            sum(gb.total_owed) as total_owed,
            -sum(COALESCE(gb2.total_owed, 0)) as inverse_owed,
            sum(gb.total_owed - COALESCE(gb2.total_owed, 0)) as net_owed
        FROM group_balances gb
        LEFT JOIN group_balances gb2 
            on gb2.group_id = gb.group_id
            and gb2.creditor = gb.debtor
            and gb2.debtor = gb.creditor
        GROUP BY gb.group_id, gb.creditor, gb.debtor
        HAVING sum(gb.total_owed - COALESCE(gb2.total_owed, 0)) > 0
    ), gbs as (
        SELECT 
            gbc.group_id,
            gbc.creditor,
            gbc.debtor,
            gbc.net_owed,
            gbc2.creditor as creditor_owes,
            COALESCE(gbc2.net_owed, 0) as chain_creditor_owed,
            gbc.net_owed - COALESCE(gbc2.net_owed, 0) as new_net_owed
        FROM gbc
        LEFT JOIN gbc gbc2 on gbc2.group_id = gbc.group_id and gbc2.debtor = gbc.creditor
    ), ua as (
        SELECT 
            group_id,
            creditor,
            debtor,
            new_net_owed as owes
        FROM gbs
        UNION ALL
        SELECT 
            group_id,
            creditor_owes as creditor,
            debtor,
            chain_creditor_owed as owes
        FROM gbs
        WHERE creditor_owes is not null
        UNION ALL
        SELECT
            group_id,
            creditor_owes as creditor,
            creditor as debtor,
            -chain_creditor_owed as owes
        FROM gbs
        WHERE creditor_owes is not null
    )
    SELECT
        group_id,
        creditor,
        debtor,
        sum(owes) as total_owed
    FROM ua
    GROUP BY group_id, creditor, debtor
    HAVING sum(owes) > 0
    ORDER BY group_id, creditor, debtor
);

CREATE VIEW group_balances_net AS (
	SELECT
		tx.group_id,
		s.split_user as user,
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
		) AS net_balance
	FROM splits s
	JOIN transactions tx ON s.transaction_id = tx.id
	GROUP BY tx.group_id, s.split_user
	ORDER BY tx.group_id, s.split_user
);
