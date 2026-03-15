# ghfs

[![Go Reference](https://pkg.go.dev/badge/github.com/unstoppablemango/ihfs/ghfs.svg)](https://pkg.go.dev/github.com/unstoppablemango/ihfs/ghfs)
[![Go version](https://img.shields.io/github/go-mod/go-version/UnstoppableMango/ihfs?filename=ghfs/go.mod)](https://github.com/UnstoppableMango/ihfs/blob/main/ghfs/go.mod)
[![Codecov](https://img.shields.io/codecov/c/github/UnstoppableMango/ihfs?flag=ghfs&label=coverage)](https://codecov.io/gh/UnstoppableMango/ihfs?flags[0]=ghfs)

An `io/fs` implementation for the GitHub API.

## Usage

### Creating a filesystem

```go
// Unauthenticated (public repos only, rate-limited)
fsys := ghfs.New()

// Authenticated
fsys := ghfs.New(ghfs.WithAuthToken(os.Getenv("GITHUB_TOKEN")))
```

### Opening files

`Open` accepts GitHub web URLs, API URLs, schemeless host prefixes, or raw API paths:

```go
// GitHub web URL
f, err := fsys.Open("https://github.com/owner/repo/blob/main/README.md")

// Schemeless shorthand
f, err := fsys.Open("github.com/owner/repo/blob/main/README.md")

// Raw GitHub API path
f, err := fsys.Open("repos/owner/repo/contents/README.md?ref=main")

// raw.githubusercontent.com URL
f, err := fsys.Open("https://raw.githubusercontent.com/owner/repo/main/README.md")
```

Release assets are also supported:

```go
f, err := fsys.Open("github.com/owner/repo/releases/download/v1.0.0/binary.tar.gz")
```

### Typed helpers

The `util.go` helpers decode API responses directly into go-github types:

```go
user, err := ghfs.OpenOwner(fsys, "owner")

repo, err := ghfs.OpenRepository(fsys, "owner", "repo")

branch, err := ghfs.OpenBranch(fsys, "owner", "repo", "main")

release, err := ghfs.OpenRelease(fsys, "owner", "repo", "v1.0.0")

content, err := ghfs.OpenContent(fsys, "owner", "repo", "main", "README.md")
```

### Using with `io/fs`

`ghfs.Fs` satisfies `fs.FS`, so standard library functions work:

```go
fsys := ghfs.New(ghfs.WithAuthToken(token))

data, err := fs.ReadFile(fsys, "repos/owner/repo/contents/file.txt?ref=main")
```

### Context and custom HTTP clients

```go
// Custom HTTP client (e.g. for testing or proxies)
fsys := ghfs.New(ghfs.WithHttpClient(myHTTPClient))

// Custom context per operation
fsys := ghfs.New(ghfs.WithContextFunc(func(f *ghfs.Fs, op ihfs.Operation) context.Context {
    return myCtx
}))
```

## Other Projects

- ghfs: <https://github.com/k1LoW/ghfs>
- ghfs: <https://github.com/johejo/ghfs>
- aferox: <https://github.com/unmango/aferox> (shameless self-plug)
