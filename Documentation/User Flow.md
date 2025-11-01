# User Flow

## MVP Endpoints (RESTful Structure)

1. **Sign up**
    - `POST /users`
    
2. **Sign in** (To be implemented later)
    - Authentication endpoint TBD
    
3. **Create group**
    - `POST /groups`
    - `GET /groups` - List all groups
    - `GET /groups/{id}` - Get specific group
    - `PUT/PATCH /groups/{id}` - Update group
    - `DELETE /groups/{id}` - Delete group
    
4. **Add members to group**
    - `POST /groups/{group_id}/members`
    - `GET /groups/{group_id}/members` - List group members
    
5. **Add transactions to group**
    - `POST /groups/{group_id}/transactions`
    - `GET /groups/{group_id}/transactions?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD` - List group transactions in period
    
6. **Split transactions with group members**
    - `POST /transactions/{transaction_id}/splits`
    - `GET /transactions/{transaction_id}/splits` - List splits for transaction
    - Transactions and splits are created separately for flexibility
    
7. **See amounts owed / true up reports**
    - `GET /groups/{group_id}/balances` - Get all balance views (detailed, net, simplified)
    - Returns three balance types:
        - `balances` - Detailed pairwise debts
        - `net_balances` - Net position per member
        - `simplified_owes` - Optimized settlement paths
    
8. **Record true up transaction**
    - `POST /groups/{group_id}/transactions` - Same as step 5, with appropriate transaction type

## Additional Direct Access Endpoints

- `GET/POST/PUT/DELETE /users/{id}` - User management
- `GET/POST/PUT/DELETE /transactions/{id}` - Direct transaction access  
- `GET/POST/PUT/DELETE /splits/{id}` - Direct split access
- `GET /users/{user_id}/transactions` - Transactions by user (RESTful nested route)
- `GET /splits/user/{user_id}` - Splits by user (direct access)
- `GET /splits/transaction/{transaction_id}` - Splits by transaction (direct access, prefer `/transactions/{id}/splits`)

## Notes

- Nested resource endpoints follow RESTful principles (parent/child relationships)
- Date parameters for transactions use ISO 8601 format (YYYY-MM-DD)
- All endpoints support pagination via `limit` and `offset` query parameters
- Balance endpoint provides multiple views for different reporting needs
