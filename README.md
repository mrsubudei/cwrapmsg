# cwrapmsg

cwrapmsg checks that the message of a wrapped error matches the name of the function or method which returned the error.

It finds all calls of `errors.Wrap` / `errors.Wrapf` from `github.com/pkg/errors` and `fmt.Errorf` with `%w`
and compares the message with the name of the called function. Mismatches are printed as `file:line:column`.

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

Run it from the root directory of the project:

```sh
cwrapmsg
```

| Flag | Description |
|---|---|
| `-u`, `--unstaged` | check only changed and untracked files (from `git status`) |
| `-c`, `--cyrillic` | check messages with cyrillic symbols, they are skipped by default |
| `-s`, `--skip` | do not use the exclusion list from `cwrapmsg.txt` |
| `-h`, `--help` | show help information |
| `-v`, `--version` | print the version |

Directories `.git` and `.cache` and generated files (`*.pb.*`, `*_gen.go`) are skipped.

The exit code is always 0, even if mismatches are found. To fail a CI job, check that the output is empty.

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

return errors.Wrap(err, "s.store.getbyids") // the comparison is case-insensitive

return errors.Wrap(err, "store.get_by_ids") // underscores are ignored
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

- Messages with cyrillic symbols, unless the `-c` flag is set.
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
variable above it. These rare cases are not reported, examples are in [testdata/not_found](testdata/not_found):

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

- A wrap inside another wrap which is not checked (cyrillic message, message is not a string literal,
  `fmt.Errorf` without `%w`):

  ```go
  return errors.Wrap(errors.Wrap(foo(), "bar"), "ошибка")
  ```

## Exclusions

Exclusions do not work for wraps where the error is a direct function call, such wraps are always checked:

```go
//nolint:cwrapmsg
return errors.Wrap(foo(), "bar") // reported anyway
```

### nolint

Add the `//nolint:cwrapmsg` comment on the line above the wrap or on the same line.
The comment must be exactly `//nolint:cwrapmsg` (several linters separated by commas are allowed: `//nolint:lll,cwrapmsg`),
an explanation after it like `//nolint:cwrapmsg // reason` is not supported.

```go
//nolint:cwrapmsg
return errors.Wrap(err, "error while get by ids")
```

```go
return errors.Wrap(err, "error while get by ids") //nolint:cwrapmsg
```

### cwrapmsg.txt

Packages, files, or functions and methods in a file can be excluded with the `cwrapmsg.txt` file
in one of the directories: `.idea`, `.vscode` or `scripts`.

One entry per line:

- a package or a file: the path relative to the root of the project;
- functions in a file: the path, a space and names separated by commas without spaces;
- a method: the name of the type, an underscore and the name of the method.

Names are case-sensitive. The file must not contain empty lines: everything after an empty line is ignored.

```
internal/app
internal/config/config.go
internal/config/db.go GetDBConfig,GetClickHouseConfig
internal/client/cdn/client.go ClientImpl_PutObject
```
