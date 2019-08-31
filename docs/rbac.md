# RBAC Documentation

RBAC query:

```sql
SELECT rp
FROM user_roles ur, role_permissions rp
WHERE ur.user_id = <user ID> AND ur.role_id = rp.role_id AND '<method + path>' ~ rp.permission
LIMIT 1;
```
