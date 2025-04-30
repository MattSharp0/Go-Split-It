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

