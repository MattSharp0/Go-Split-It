# ToDo's

- Review endpoints and filters
    - ~~Date filters should be optional for transactions by group & user~~
- Troubleshoot balances endpoint reports missing user data
- ~~Generate transaction for adding all splits for a transaction at once~~
    - Add function to calculte % or $ amount if missing from API request 
- Implement logging middleware for all endpoints
    - Debug for requests and server connection
    - Info for startup and shutdown
    - Configurable levels for
- Build out testing
    - Warning / Fatal for errors
- Add user authentication & scope
- Add 'created by' timestamp to transactions, splits, and groups
- Add config file for setup
- Add start and end date for group members
    - allows for transactions to apply to only those active at tx date
- Generate required dockerfile(s)
