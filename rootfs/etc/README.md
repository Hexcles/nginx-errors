## mime.types

This is a minimal MIME database (see `man 5 mime.types`) derived from
https://releases.pagure.org/mailcap/mailcap-2.1.53.tar.xz

The file is processed in the following ways as we only use it for
[`mime.ExtensionsByType`](https://pkg.go.dev/mime#ExtensionsByType):

1.  Skip lines without any extension.
2.  Only keep the first, canonical extension; otherwise, `mime.ExtensionsByType`
    would sort the extensions and we can no longer tell which one is canonical.
3.  Drop comments.

```bash
awk '/^[^#]/ && NF>1 {print $1, $2}' mime.types
```
