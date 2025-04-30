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