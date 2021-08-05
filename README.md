# pgtypes JSON demo

```
$ go run main.go "postgresql://postgres:test@localhost:5432/dune?sslmode=disable"
Marshalling result of `SELECT '{1,2}'::text[]` into JSON.
pgtype: {"Elements":["1","2"],"Dimensions":[{"Length":2,"LowerBound":1}],"Status":2}
custom: ["1","2"]
```
