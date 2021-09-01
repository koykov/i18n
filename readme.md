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

## Pluralization

i18n supports plural formulas. Default formula has format `"<singular>|<plural>"` and supports two ranges: `[0, 1]` for
singular and `[2, +Inf]` for plural.

In addition to default formulas i18n supports extended formats: `"[low,high] translation|..."`, `"{exact} translation|..."`
and various combination of them.

Let's pluralize for example [enemy army counts](https://heroes.thelazy.net/index.php/Creature) for Heroes III game:
```go
db.Set("en.h3.army_size", "[1,5] Few|[5,10] Several|[10,20] Pack|[20,50] Lots|[50,100] Horde|[100,250] Throng|[250,500] Swarm|[500,1000] Zounds|[1000,*] Legion")
db.GetPlural("en.h3.army_size", "N/D", 0) // N/D
db.GetPlural("en.h3.army_size", "", 2) // Few
db.GetPlural("en.h3.army_size", "", 19) // Pack
db.GetPlural("en.h3.army_size", "", 20) // Lots
db.GetPlural("en.h3.army_size", "", 333) // Swarm
db.GetPlural("en.h3.army_size", "", 999) // Zounds
db.GetPlural("en.h3.army_size", "", 1e9) // Legion
```

Check [i18n_test.go](i18n_test.go) to see these examples in action.

## Transaction support

To reduce lock pressure you may use transaction. See [txn_test.go](txn_test.go) for example.
