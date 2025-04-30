DELETE FROM splits;
DELETE FROM transactions;
DELETE FROM group_members;
DELETE FROM groups;
DELETE FROM users;
ALTER SEQUENCE splits_id_seq RESTART;
ALTER SEQUENCE transactions_id_seq RESTART;
ALTER SEQUENCE group_members_id_seq RESTART;
ALTER SEQUENCE groups_id_seq RESTART;
ALTER SEQUENCE users_id_seq RESTART;

INSERT INTO "users" (name) 
VALUES ('Alice'), ('Bob'), ('Charlie'), ('Dan')
RETURNING *;

INSERT INTO "groups" (name)
VALUES ('Test Group 1'), ('Test Group 2'), ('Test Group 3')
RETURNING *;

INSERT INTO "group_members" (group_id, user_id)
VALUES 
	(1,1), (1,2), (1,3), (1,4), 
	(2,1), (2,2), 
	(3,1), (3,2), (3,3)
RETURNING *;

INSERT INTO transactions (name, amount, by_user, group_id, category, transaction_date)
VALUES 
    ('test-1',500,1,1,'Travel',CURRENT_DATE), -- Group 1
    ('test-2',20,2,1,'Dining Out',CURRENT_DATE), -- Group 1
    ('test-3',80,1,1,'Dining Out',CURRENT_DATE), -- Group 1
    ('test-4',200,2,1,'Groceries',CURRENT_DATE), -- Group 1
    ('test-5',300,3,1,'Shopping',CURRENT_DATE), -- Group 1
	('test-6',225,4,1,'Groceries',CURRENT_DATE), -- Group 1
	('test-7',120,4,1,'Dining Out',CURRENT_DATE), -- Group 1
	
	('test-8',225,1,2,'Groceries',CURRENT_DATE), -- Group 2
	('test-9',120,2,2,'Dining Out',CURRENT_DATE), -- Group 2
	
	('test-10',10,1,3,'',CURRENT_DATE), -- Group 3
	('test-11',10,2,3,'',CURRENT_DATE), -- Group 3
	('test-12',10,3,3,'',CURRENT_DATE) -- Group 3
RETURNING *;

INSERT INTO splits (transaction_id, tx_amount, split_percent, split_amount, split_user)
VALUES 
	(1,500,.5, 250, 1), -- Group 1
	(1,500,.5, 250, 2), -- Group 1
	(2,20,.5, 10, 1), -- Group 1
	(2,20,.5, 10, 2), -- Group 1
	(3,80,.5, 40, 1), -- Group 1
	(3,80,.5, 40, 2), -- Group 1
	(4,200,.5, 100, 1), -- Group 1
	(4,200,.5, 100, 2), -- Group 1
	(5,300,.333333, 100, 1), -- Group 1
	(5,300,.333333, 100, 2), -- Group 1
	(5,300,.333333, 100, 3), -- Group 1
	(6,225,.4, 90, 4), -- Group 1
	(6,225,.6, 135, 3), -- Group 1
	(7,120,.25, 30, 1), -- Group 1
	(7,120,.25, 30, 2), -- Group 1
	(7,120,.25, 30, 3), -- Group 1
	(7,120,.25, 30, 4), -- Group 1
	
	(8,225,.4, 90, 1), -- Group 2
	(8,225,.6, 135, 2), -- Group 2
	(9,120,.5, 60, 1), -- Group 2
	(9,120,.5, 60, 2), -- Group 2
	
	(10,10,1,5,2), -- Group 3
	(10,10,1,5,3), -- Group 3
	(11,10,1,5,1), -- Group 3
	(11,10,1,5,3), -- Group 3
	(12,10,1,5,1), -- Group 3
	(12,10,1,5,2) -- Group 3
RETURNING *;
