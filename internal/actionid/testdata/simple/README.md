## generation

```
$ GODEBUG=gocachehash=1 go build -p 1 -trimpath -o /dev/null . 2>&1 | grep 'HASH\[build ' > go1.21.0-drawin-amd64.hash
```