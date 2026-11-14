# cwrapmsg

cwrapmsg checks that the message of a wrapped error matches the name of the function or method which returned the error.

It finds all calls of `errors.Wrap` / `errors.Wrapf` from `github.com/pkg/errors` and `fmt.Errorf` with `%w`
and compares the message with the name of the called function.

The linter is built on [`go/analysis`](https://pkg.go.dev/golang.org/x/tools/go/analysis).

```go
items, err := s.store.GetByIDs(ctx, ids)
if err != nil {
	return errors.Wrap(err, "store.GetByID") // reported: the message does not match GetByIDs
}
```

Such messages make the error chain read like a stack trace (`handler: service.GetItems: store.GetByIDs: no rows`),
and the linter catches messages which became outdated after renaming or copy-paste.

## Installation

```sh
go install github.com/mrsubudei/cwrapmsg/cmd/cwrapmsg@latest
```

## Usage

Run it in the root directory of the module with the packages to check:

```sh
cwrapmsg ./...
```

Or with `go vet`:

```sh
go vet -vettool=$(which cwrapmsg) ./...
```

Mismatches are reported as:

```
internal/store/store.go:42:10: message in errors.Wrap does not match function or method name
```

If mismatches are found, the exit code is not zero.

| Flag | Description |
|---|---|
| `-ignore-cyrillic` | skip messages with cyrillic symbols, they are checked by default |
| `-case-insensitive` | compare the message with the name case-insensitively |

The other flags are the standard flags of `go/analysis` linters (`-json`, `-test`, etc.), see `cwrapmsg -h`.

Generated files are skipped: files with the `// Code generated ... DO NOT EDIT.` comment,
`*.pb.*` and `*_gen.go` files.

## Message format

For the call

```go
items, err := s.store.GetByIDs(ctx, ids)
```

all of these messages are correct:

```go
return errors.Wrap(err, "s.store.GetByIDs") // the full name

return errors.Wrap(err, "store.GetByIDs") // a part of the name, without "s."

return errors.Wrap(err, "GetByIDs") // a part of the name, without "s.store."

return errors.Wrap(err, "store.Get_By_IDs") // underscores are ignored
```

The comparison is case-sensitive. With the `-case-insensitive` flag these messages are correct too:

```go
return errors.Wrap(err, "s.store.getbyids")

return errors.Wrap(err, "store.get_by_ids")
```

The same rules apply to `fmt.Errorf`:

```go
return fmt.Errorf("store.GetByIDs: %w", err)

return fmt.Errorf("store.GetByIDs: %w, name: %v", err, name) // other verbs are allowed
```

Only the beginning of the message is compared: letters, digits and dots up to the first other symbol.
So details can be added after the name:

```go
err := processByBucket(ctx, bucket)

return errors.Wrapf(err, "processByBucket: %d", bucket) // only "processByBucket" is compared
```

Because of this the linter does not fit projects where messages are written in another format,
they will be reported:

```go
return errors.Wrap(err, "store::GetByIDs") // only "store" is compared

return errors.Wrap(err, "store GetByIDs") // only "store" is compared

return errors.Wrap(err, "error while get by ids") // only "error" is compared
```

If methods are called in a chain, only the name of the last method is compared:

```go
err := c.readConn(ctx).Execx(ctx, qb)

return errors.Wrap(err, "Execx") // only "Execx" is compared
```

## What is not checked

- Messages with cyrillic symbols, if the `-ignore-cyrillic` flag is set.
- Messages which are not string literals: `errors.Wrap(err, msg)`.
- `fmt.Errorf` without `%w` or with more than one `%w`.
- Wraps where the error was returned in another function or closure than the wrap:

  ```go
  func MyFunc() error {
  	var err error
  	var result int

  	someFuncOnce.Do(func() {
  		result, err = somePkg.Method() // inside the closure
  	})

  	return errors.Wrap(err, "somePkg.Method") // outside the closure: not checked
  }
  ```

## Known limitations

The analysis is based on the source code only and compares the wrap with the last assignment of the error
variable above it. These rare cases are not reported,
examples are in [testdata/cwrapmsg/limitations](testdata/cwrapmsg/limitations):

- The call and the wrap are on the same line:

  ```go
  if err := foo(); err != nil { return errors.Wrap(err, "bar") }
  ```

- `fmt.Errorf` with flags, width, precision or explicit argument index in verbs before `%w`
  (`%+v`, `%5d`, `%.2f`, `%[2]w`): the argument for `%w` is determined incorrectly.

  ```go
  return fmt.Errorf("bar %+v: %w", x, foo())
  ```

- A variable with the same name declared in an inner scope (shadowing):

  ```go
  _, err := load()
  if err := set(); err != nil {
  	log.Println(err)
  }

  if err != nil {
  	return errors.Wrap(err, "set") // err is from load, but it is compared with set
  }
  ```

- Several calls on the right side of an assignment:

  ```go
  err, n := foo(), bar()

  return errors.Wrap(err, "bar") // err is from foo, but it is compared with bar
  ```

- A wrap inside another wrap which is not checked (cyrillic message with `-ignore-cyrillic`,
  message is not a string literal, `fmt.Errorf` without `%w`):

  ```go
  return errors.Wrap(errors.Wrap(foo(), "bar"), "ошибка")
  ```

## Exclusions

The linter has no exclusions of its own: they are provided by the tool which runs the analyzer,
for example `//nolint:cwrapmsg` comments and exclusion rules in golangci-lint.
