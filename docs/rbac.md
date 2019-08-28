# RBAC Documentation

RBAC query:

```
select
	id
from user_roles as ur
	inner join role_permissions as rp on ur.role_id = rp.role_id
where
	ur.id = 3 and 'cagar' ~ rp.permission
limit 1;
```