# nginx-errors

![docker-build](https://github.com/Hexcles/nginx-errors/actions/workflows/docker-build.yml/badge.svg)
![golangci-lint](https://github.com/Hexcles/nginx-errors/actions/workflows/golangci-lint.yml/badge.svg)

This is a minimal and robust [custom error backend][custom-error] for [NGINX
Ingress Controller][ingress-nginx] based on the official [example][example].

## Example

First, turn on [custom-http-errors][custom-http-errors] in your controller,
e.g. `custom-http-errors: 404,494,500,529`.

```dockerfile
FROM hexcles/nginx-errors:latest

# Map some non-standard status codes to their standard counterparts.
ENV STATUS_CODE_MAPPING="494:400,529:503"

# Customize HTML response for 404.
COPY 404.html /www/404.html

# Customize JSON response for 5xx.
COPY 5xx.json /www/5xx.json
```

## Design

The main design goal is minimalism, through which we also achieve robustness
(the priority is in that order). It uses *only* Go standard libraries,
effectively a bare `http.ListenAndServe` at its core.

The binary is statically built with CGO disabled. The Docker image is also
minimal (built from `scratch`) without a shell or even the usual [FHS][fhs].

## Customization

Customization is done through building another Docker image on top of this one,
where you can set environment variables to configure behaviours and/or overlay
some files to customize the error responses.

### Custom error responses

You can put `[code].[ext]` in `/www` of the container to customize the responses for
certain status codes and `Accept`ed MIME types as requested by the client.

*   The `[code]` portion of the file name can be either a specific status code
    (e.g. `404`), or a range of codes in the form of e.g. `4xx`.
*   The `[ext]` portion of the file name should correspond to the included
    [`/etc/mime.types`][mime.types] (you can also overlay this file in your
    Docker image if you need a MIME type that is not included).

The image includes default error responses for 404, 4xx, 500 and 5xx in HTML
and JSON respectively.

### `ENV` Configurations

*   `DEBUG`: turn on debug logging and response headers. (Do *NOT* use in
    production.)
*   `DEFAULT_RESPONSE_FORMAT`: set the default response format when the
    there is no requested MIME type or it cannot be recognized. Default to
    `text/html`.
*   `STATUS_CODE_MAPPING`: in the format of `SRC_A:DST_A,SRC_B:DST_B,...`
    mapping source status codes (returned by your backend application) to
    destination status codes (seen by the client). Note that you should put the
    *source* status codes in `custom-http-errors` while the *destination* status
    codes will be used to look up error responses.

[custom-error]: https://kubernetes.github.io/ingress-nginx/user-guide/custom-errors/
[ingress-nginx]: https://kubernetes.github.io/ingress-nginx/
[example]: https://github.com/kubernetes/ingress-nginx/tree/main/images/custom-error-pages
[custom-http-errors]: https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#custom-http-errors
[fhs]: https://en.wikipedia.org/wiki/Filesystem_Hierarchy_Standard
[mime.types]: ./etc/mime.types
