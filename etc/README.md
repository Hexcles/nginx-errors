## mime.types

This is a minimal MIME database (see `man 5 mime.types`) derived from
https://releases.pagure.org/mailcap/mailcap-2.1.53.tar.xz

The file is processed in the following ways as we only use it to get extensions
by MIME types:

1.  Skip lines without any extension.
2.  Only keep the first, canonical extension.
3.  Drop comments.

```bash
awk '/^[^#]/ && NF>1 {print $1, $2}' mime.types
```
