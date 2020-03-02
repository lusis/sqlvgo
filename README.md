# sqlvgo

The purpose of this repo is because I got nerdsniped (unintentionally).
The general statement was "filtering in SQL is always faster".

The thing I was curious about was challenging that conventional wisdom.
As with all things that's true in the general case. This is what databases are designed to do right?

## so why not just filter in sql?

The problem here isn't so much that filtering in sql is hard.

`select * from foo where y in (1,3,5,7) AND z IN (1,5,7,9)`

That's fairly simple.

The problem is that filtering in sql in Go is painful and with MySQL it's exponentially painful.

## sql and go

First off if we want to duplicate the above query in Go using `database/sql`, we can't.
Because go is a strongly typed language, we have to do a few things:

- explicitly name our columns
- map those columns to a type
- convert predicates to using placeholders
- provide params to replace those placeholders

so the above query becomes:

`select a,b,c,x,y,z from foo where y in (?,?,?,?) AND z IN (?,?,?,?)`

and params is now:

`[]interface{1,3,5,7,1,5,7,9}`

Okay that's not so bad right? Sure it's painful but doable.

A few gotchas:

- placeholder count and param count need to match (makes sense)
- param order must match placeholder order (...okay makes sense)

so essentially your go code to perform the above becomes something like this:

```go
type foo struct {
    A string
    B time.Time
    C time.Time
    X int
    Y int
    Z string
}

q := `select a,b,c,x,y,z from foo where y in (?,?,?,?) AND z IN (?,?,?,?)`
params := []interface{1,3,5,7,1,5,7,9}
stmt, err := db.Prepare(query)
if err != nil {
    return nil, err
}
defer stmt.Close()
rows, err := stmt.Query(params...)
if err != nil {
    return nil, err
}
defer rows.Close()

results := make([]foo, 0)
for rows.Next() {
    var f Foo
    if err := rows.Scan(
        &f.A,
        &f.B,
        &f.C,
        &f.X,
        &f.Y,
        &f.Z,
    ); err != nil {
        rows.Close() // nolint: errcheck
        return nil, err
    }
    results = append(results, f})
}
if err := rows.Err(); err != nil {
    rows.Close()
    return nil, err
}
rows.Close() // nolint: errcheck
```

![escalated](https://www.dictionary.com/e/wp-content/uploads/2018/03/that-escalated-quickly.jpg)

And it's worth noting that your code above ONLY works if you know exactly how many placeholders and params you'll need.
If you accept variable input, you now need to do some variant of the following:

```go
placeholders := make([]string, 0)
params := make([]interface{}, 0)
// iterate once for X
for _, i := range ints {
    params = append(params, i)
}
// iterate again for Y
for _, i := range ints {
    params = append(params, i)
}
for i := 1; i <= len(ints); i++ {
    placeholders = append(placeholders, "?")
}
query := "select a,b,c,x,y,z from foo where"
query = query + "x IN (" + strings.Join(placeholders, ",") + ")"
query = query + " AND "
query = query + "y IN (" + strings.Join(placeholders, ",") + ")"
stmt, _ := db.Prepare(query)
rows, _ := stmt.Query(params...)
```

Note that in this simplified example, I'm using the same values for each `IN` clause for example purposes.

## Our example

Imagine you have a grpc service that has a message type like so:

```proto
message FindFoosRequest {
    repeated State states = 1; // this is a state in terms of lifecycle not geography
    repeated Type types = 2;
    repeated string ids = 3;
    google.protobuf.Timestamp before = 4;
    google.protobuf.Timestamp after = 5;
}
```

If you're not familiar with protobufs, this is the message that would be sent by a grpc client.
In proto3, all fields are optional.
Clients can provide pretty much any permutation of values there.
To build this into a query you would need to map each field to a predicate:

- `repeated State states` -> `where state in (?,?,.....)`
- `repeated Type types` -> `where type in (?,?,.....)`
- `repeated string ids` -> `where id in (?,?,.....)`
- `before` -> `where created_at <= ?`
- `after` -> `where created_at >= ?`

You can see this gets cumbersome to convert convert into the final SQL statement...

Filtering in go in this case would be much simpler if you just got all the records and iterated base on the zero values of each message field.

## About the benchmarks

I've made the code SUPER modular because I was reusing a good chunk of it in different parts. I wouldn't normally write it this way.
When it comes to the go filtering, I've done things based on constraints I'm implementing for the tests not how I normally would.
There are a bunch of different techniques in go for optimizing iterations of slices and filtering out elements.

### Table model

```sql
CREATE TABLE `testdata` (
`counter` INT UNSIGNED NOT NULL AUTO_INCREMENT,
`id` VARCHAR(191) NOT NULL,
`name` VARCHAR(191) NOT NULL,
`rtype` TINYINT UNSIGNED NOT NULL,
`rstate` TINYINT UNSIGNED NOT NULL,
`created_at` DATETIME(6) NOT NULL DEFAULT NOW(6),
`updated_at` DATETIME(6) NOT NULL DEFAULT NOW(6) ON UPDATE NOW(6),
PRIMARY KEY (`counter`),
UNIQUE INDEX `id` (`id`),
INDEX `state` (`rstate`),
INDEX `type` (`rtype`),
INDEX `state_and_type` (`rstate`,`rtype`)
)
CHARSET='utf8mb4'
ENGINE=InnoDB;
```

This roughly aligns with the use case I described above.

In this data model we're using UUIDs as identifiers and an autoincrement int primary key due to how innodb handles indexes in relation to the primary key.
Long story short, innodb prefixes part of the primary key (8 bytes or something) to all indexes.
When you use a UUID as your primary key, it's SLOW. The general "trick" is to use a standard int primary key that you don't give a shit about.


### Actual benchmarks

All benchmarks except the 100k row follow a similar pattern:

- for some number of records, populate the table with random data
- run one of the filtering strategies (go vs sql)

The number of records is defined at the top of the benchmark testfile: `var testRecCounts = []int{50, 100, 1000, 5000, 10000, 50000, 100000}`

You'll need mysql running (`scripts/start-local-mysql`) before running the benchmarks. Everything is pretty much hardcoded to talk to the docker mysql instance (5.7)

Additionally, I ran all benchmarks with a `-benchtime=100x` (sometimes more depending). Go's benchmark varies the number of iterations per run and I wanted a bit more consistency.
For instance, when you started hitting larger numbers of rows, you would get fewer iterations sampled per filtering type.

#### `BenchmarkSelectAll`/`BenchmarkSelectAllIndexed`

There are two variants here based on the innodb scenario I mentioned above.
I wanted to get some numbers on selecting everything vs selecting everything and ordering by the pk.

you can run the benchmarks like so:

`CGO_ENABLED=0 go test -run=none -v -bench=BenchmarkSelectAll -benchmem -failfast -parallel=1 -timeout=2h -memprofile=mem-all.out -cpuprofile=cpu-all.out -benchtime=100x`

#### `BenchmarkSelectSomeIndexed`

This is where we test the following three mechanisms for filtering:

- filtering in sql with a naive query builder as described above
- filtering in sql using [goqu](https://github.com/doug-martin/goqu) as the query builder
- selecting everything and filtering in go

I've tried to do some things that make sense.

- only time the actual call to the database
- ensure mysql is flushed between each iteration
- ensure that functional calls don't get optimized away


There's probably a mistake or 12 in there but it's the same mistake for each permutation so I guess it's apples to apples?

you can run the benchmarks like so:

`CGO_ENABLED=0 go test -run=none -v -bench=BenchmarkSelectSome -benchmem -failfast -parallel=1 -timeout=2h -memprofile=mem-some.out -cpuprofile=cpu-some.out -benchtime=100x`

#### `Benchmark100kSelectSomeIndexed`

This is the benchmark that will not prepopulate the table randomly each time. For this I wanted more determinism so I wanted to work with the same dataset over each `go test -bench` call.

You can prepopulate the db with:

`NUM_RECORDS=100000 go run ./cmd/populate/main.go`

You can see the final distribution of records via sql:

`select rstate, rtype, count(*) as count from testdata group by rstate, rtype;`

You can run this benchmark like so:

`CGO_ENABLED=0 go test -run=none -v -bench=Benchmark100k -benchmem -failfast -parallel=1 -timeout=2h -memprofile=mem-big.out -cpuprofile=cpu-big.out -benchtime=100x`

