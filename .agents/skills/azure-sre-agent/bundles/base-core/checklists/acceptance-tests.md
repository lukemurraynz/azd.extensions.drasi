# Base-Core Acceptance Tests

Run these checks before enabling production automation.

1. Historical incident test routes to expected diagnostic agent.
2. Diagnostic output includes UTC evidence and uncertainty statement.
3. Remediation proposal includes blast-radius and rollback.
4. Remediation approval hook blocks execution without approval.
5. Daily health task runs manually and exits cleanly.
6. Notification channel sends structured update successfully.
