-- over_limit_since is when the app was first seen with more enabled commands
-- or event listeners than its plan allows. The newest ones are disabled after
-- a grace period, and it's cleared once the app is under the limits again.
ALTER TABLE apps ADD COLUMN over_limit_since TIMESTAMP;
