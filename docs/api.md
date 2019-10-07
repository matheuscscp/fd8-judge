# The RBAC API
In our RBAC authorization system, each role is an API resource (of the `role`
resource type) associated to a set of regex permissions responsible for recognizing
string representations of (method, URI) tuples. The query used to authorize a
user for an endpoint is the following:

```sql
SELECT rp
FROM user_roles ur, role_permissions rp
WHERE ur.user_id = <user ID> AND ur.role_id = rp.role_id AND '<method + path>' ~ rp.permission
LIMIT 1;
```

Examples below show some endpoints
and their respective possible regex permissions. Within the examples context, `%publicID`
represents the POSIX regex `[a-z0-9](-[a-z0-9])*`, which defines our set of possible public IDs.

### Example 1
In the example below, the regex permission could be granted to the role
`staff-in-my-contest`, used to access sub-resources of the `contest`
resource publicly identified by `my-contest`.
```
endpoint 1: PUT /contests/:contestPublicID/balloon-tasks/:taskPublicID/self-assign
endpoint 2: PUT /contests/:contestPublicID/balloon-tasks/:taskPublicID/complete
endpoint 3: PUT /contests/:contestPublicID/print-tasks/:taskPublicID/self-assign
endpoint 4: PUT /contests/:contestPublicID/print-tasks/:taskPublicID/complete
endpoint 5: DELETE /contests/:contestPublicID/balloon-tasks/:taskPublicID/self-assign
endpoint 6: DELETE /contests/:contestPublicID/balloon-tasks/:taskPublicID/complete
endpoint 7: DELETE /contests/:contestPublicID/print-tasks/:taskPublicID/self-assign
endpoint 8: DELETE /contests/:contestPublicID/print-tasks/:taskPublicID/complete
```
```
regex: ^(PUT|DELETE)/contests/my-contest/(balloon|print)-tasks/%publicID/(self-assign|complete)$
```

### Example 2
In the example below, the regex permission could be granted to the role
`judge-in-my-contest`, used to access sub-resources of the `contest`
resource publicly identified by `my-contest`.
```
endpoint 1: POST /contests/:contestPublicID/judgements/:solutionPublicID
endpoint 2: PUT /contests/:contestPublicID/judgements/:solutionPublicID
endpoint 3: DELETE /contests/:contestPublicID/judgements/:solutionPublicID
```
```
regex: ^(POST|PUT|DELETE)/contests/my-contest/judgements/%publicID$
```

## API endpoints
The *main endpoints* subset of the RBAC API provides full control over the RBAC system.

To provide easy-to-use, fine-grained RBAC  management we also implement some extra API
subsets bound to some important domains of the application. These subsets are:
* *User permissions*: A subset to manage individual permissions. Individual permissions
are possible because the application creates a dedicated role for each user.
* *Programing contests*: A subset to manage RBAC for endpoints referring to a specific
`contest` resource.
* *Organizations*: A subset to manage RBAC for endpoints referring to a specific
`organization` resource.

### Main endpoints
The RBAC API main endpoints are listed below.

List roles:
```
GET /roles?filters=&pageSize=&page=&orderBy=
```

List permissions granted to a role:
```
GET /roles/:rolePublicID/permissions?filters=&pageSize=&page=&orderBy=
```

Grant a permission to a role:
```
POST /roles/:rolePublicID/permissions/:regexPermission
```

Revoke a permission from a role:
```
DELETE /roles/:rolePublicID/permissions/:regexPermission
```

List roles granted to a user:
```
GET /users/:userPublicID/roles?filters=&pageSize=&page=&orderBy=
```

Grant a role to a user:
```
POST /users/:userPublicID/roles/:rolePublicID
```

Revoke a role from a user:
```
DELETE /users/:userPublicID/roles/:rolePublicID
```


### User permissions
The endpoints to manage individual permissions are listed below.
The underlying role these endpoints refer to is `{:userPublicID}-permissions`,
which is bound solely to the user identified by `:userPublicID`.

List permissions granted to a user:
```
GET /users/:userPublicID/permissions?filters=&pageSize=&page=&orderBy=
```

Grant a permission to a user:
```
POST /users/:userPublicID/permissions/:regexPermission
```

Revoke a permission from a user:
```
DELETE /users/:userPublicID/permissions/:regexPermission
```

### Programming contests
The endpoints to manage RBAC for the `contest` resource type are listed below.
Here, a *contest role* is a role dedicated to grant access to a specific `contest`
resource.

List roles to access a contest:
```
GET /contests/:contestPublicID/roles?filters=&pageSize=&page=&orderBy=
```

List permissions granted to a contest role:
```
GET /contests/:contestPublicID/roles/:rolePublicID/permissions?filters=&pageSize=&page=&orderBy=
```

Grant a permission to a contest role:
```
POST /contests/:contestPublicID/roles/:rolePublicID/permissions/:regexPermission
```

Revoke a permission from a contest role:
```
DELETE /contests/:contestPublicID/roles/:rolePublicID/permissions/:regexPermission
```

List contest roles granted to a user:
```
GET /contests/:contestPublicID/users/:userPublicID/roles?filters=&pageSize=&page=&orderBy=
```

Grant a contest role to a user:
```
POST /contests/:contestPublicID/users/:userPublicID/roles/:rolePublicID
```

Revoke a contest role from a user:
```
DELETE /contests/:contestPublicID/users/:userPublicID/roles/:rolePublicID
```

### Organizations
The endpoints to manage RBAC for the `organization` resource type are listed below.
Here, an *organization role* is a role dedicated to grant access to a specific
`organization` resource.

List roles to access an organization:
```
GET /organizations/:organizationPublicID/roles?filters=&pageSize=&page=&orderBy=
```

List permissions granted to an organization role:
```
GET /organizations/:organizationPublicID/roles/:rolePublicID/permissions?filters=&pageSize=&page=&orderBy=
```

Grant a permission to an organization role:
```
POST /organizations/:organizationPublicID/roles/:rolePublicID/permissions/:regexPermission
```

Revoke a permission from an organization role:
```
DELETE /organizations/:organizationPublicID/roles/:rolePublicID/permissions/:regexPermission
```

List organization roles granted to a user:
```
GET /organizations/:organizationPublicID/users/:userPublicID/roles?filters=&pageSize=&page=&orderBy=
```

Grant an organization role to a user:
```
POST /organizations/:organizationPublicID/users/:userPublicID/roles/:rolePublicID
```

Revoke an organization role from a user:
```
DELETE /organizations/:organizationPublicID/users/:userPublicID/roles/:rolePublicID
```

# The User API

```
GET /users/:userPublicID
GET /users?filters=&pageSize=&page=&orderBy=
POST /users/:userPublicID
PUT /users/:userPublicID
DELETE /users/:userPublicID
```

# The Problem API

```
GET /problems/:problemPublicID
GET /problems?filters=&pageSize=&page=&orderBy=
POST /problems/:problemPublicID
PUT /problems/:problemPublicID
DELETE /problems/:problemPublicID

GET /solutions/:solutionPublicID
GET /solutions?filters=&pageSize=&page=&orderBy=
POST /solutions?problemPublicID=
PUT /solutions/:solutionPublicID
DELETE /solutions/:solutionPublicID

GET /solutions/:solutionPublicID/executions/:executionPublicID
GET /solutions/:solutionPublicID/executions?filters=&pageSize=&page=&orderBy=
POST /solutions/:solutionPublicID/executions
PUT /solutions/:solutionPublicID/executions/:executionPublicID
DELETE /solutions/:solutionPublicID/executions/:executionPublicID
```

# The Contest API



# The Organization API


