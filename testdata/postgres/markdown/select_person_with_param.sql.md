# Select Person With parameter Table

```sql
SELECT email, first_name FROM persons WHERE first_name=/*first_name*/'bob';
```

## Parameter

| Name       | Type   | Description |
| ---------- | ------ | ----------- |
| first_name | string | search key  |

## Test

### Case: Query Empty Test

```yaml
params: { first_name: Dan }
expect: []
```

### Case: Query Test (Failure)

```yaml
params: { first_name: Evan }
expect: []
```



