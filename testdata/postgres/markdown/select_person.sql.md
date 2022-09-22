# Select Person

This comment is ignored.

```sql
SELECT email, first_name FROM persons WHERE first_name=/*first_name*/'bob';
```

## Parameter

| Name       | Type   | Description |
|------------|--------|-------------|
| first_name | string | search key  |

## CRUD Matrix

| Table      | C | R | U | D | Description |
|------------|---|---|---|---|-------------|
| persons    | X |   |   |   |             |

## Tests

### Case: Query Evan Test

```yaml
fixtures:
  persons:
    - [employee_no, dept_no, first_name, last_name, email, created_at]
    - [4, 13, Dan, Conner, dan@example.com, 2022-09-13 10:30:15]
params: { first_name: Dan }
expect:
  - { email: dan@example.com, first_name: Dan } 
```
