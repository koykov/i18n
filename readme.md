# Internationalization

Simple i18n database.

## Usage

```go
db := i18n.New(fnv.Hasher{})
db.SetPolicy(policy.Locked)

db.Set("en.messages.welcome", "Hello there!")
db.Set("ru.messages.welcome", "Привет!")

db.SetPolicy(policy.LockFree)

fmt.Println(db.Get("en.messages.welcome"))
```

## Transaction support

To reduce lock pressure you may use transaction. See [txn_test.go](txn_test.go) for example.
