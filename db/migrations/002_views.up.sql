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
    WITH gdc as (
        SELECT
            gd.group_id,
            gd.creditor,
            gd.debtor,
            sum(gd.total_owed) as total_owed,
            -sum(COALESCE(gd2.total_owed, 0)) as inverse_owed,
            sum(gd.total_owed - COALESCE(gd2.total_owed, 0)) as net_owed
        FROM group_debts_complex gd
        LEFT JOIN group_debts_complex gd2 
            on gd2.group_id = gd.group_id
            and gd2.creditor = gd.debtor
            and gd2.debtor = gd.creditor
        GROUP BY gd.group_id, gd.creditor, gd.debtor
        HAVING sum(gd.total_owed - COALESCE(gd2.total_owed, 0)) > 0
    ), gds as (
        SELECT 
            gdc.group_id,
            gdc.creditor,
            gdc.debtor,
            gdc.net_owed,
            gdc2.creditor as creditor_owes,
            COALESCE(gdc2.net_owed, 0) as chain_creditor_owed,
            gdc.net_owed - COALESCE(gdc2.net_owed, 0) as new_net_owed
        FROM gdc
        LEFT JOIN gdc gdc2 on gdc2.group_id = gdc.group_id and gdc2.debtor = gdc.creditor
    ), ua as (
        SELECT 
            group_id,
            creditor,
            debtor,
            new_net_owed as owes
        FROM gds
        UNION ALL
        SELECT 
            group_id,
            creditor_owes as creditor,
            debtor,
            chain_creditor_owed as owes
        FROM gds
        WHERE creditor_owes is not null
        UNION ALL
        SELECT
            group_id,
            creditor_owes as creditor,
            creditor as debtor,
            -chain_creditor_owed as owes
        FROM gds
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
