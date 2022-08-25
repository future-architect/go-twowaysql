# Select Person

This comment is ignored.

```sql
SELECT email, name FROM empty_persons WHERE first_name=/*first_name*/'bob';
```

## Parameter

| Name       | Type   | Description |
|------------|--------|-------------|
| first_name | string | search key  |

## CRUD Matrix

| Table      | C | R | U | D | Description |
|------------|---|---|---|---|-------------|
| persons    | X |   |   |   |             |
