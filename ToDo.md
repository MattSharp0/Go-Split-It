# ToDo's

- Review endpoints and filters
    - Date filters should be optional for transactions by group 
- Troubleshoot balances endpoint reports missing user date
- Build out testing
- Generate transaction for adding all splits for a transaction at once
- Implement logging for all endpoints (including variables & paramaters) and variable levels
    - Debug for requests and server connection
    - Info for startup and shutdown
    - Warning / Fatal for errors
- Add user authentication & scope
- Add 'created by' timestamp to transactions, splits, and groups
- Add config file for setup
- Add start and end date for group members
    - allows for transactions to apply to only those active at tx date
- Generate required dockerfile(s)
