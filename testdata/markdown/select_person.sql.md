# Select Person

```sql
SELECT email, name FROM empty_persons WHERE first_name=/*first_name*/'bob';
```

## Test

### Case: test with alice

#### TestData

```csv :empty_persons
employee_no, email,             first_name
1,           alice@example.com, alice
2,           bob@example.com,   bob
```

#### Param

```yaml
first_name:  alice
```

#### Expect

```yaml
- email:      alice@example.com
  first_name: alice
```
