This can generate a Gofile (for https://github.com/joewalnes/go-getter) from all files (recursive) in the current directory.

# Usage
Clone the repo:

```
go get github.com/maxwellhealth/gofile-generator
```

Build it:

```
go build
```

Install it:

```
go install
```

Test it:

```
cd /path/to/some/go/package
gofile-generator
go-getter Gofile
```