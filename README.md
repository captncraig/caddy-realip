# Real-IP middleware for Caddy
Middleware for restoring real ip information when running caddy behind a proxy. Will allow other
middlewares to simply use `r.RemoteAddr` instead of decoding `X-Forwarded-For` themselves.
Analogous to nginx's [realip_module](http://nginx.org/en/docs/http/ngx_http_realip_module.html)

Checks whitelist of authorized proxy servers so we don't arbitrarily trust headers from anybody.

The real IP module is useful in situations where you are running caddy behind a proxy server.
In these situations, the actual client IP will be stored in an HTTP header, usually `X-Forwarded-For`.
The problem this creates is that other directives that rely on the client IP address, like
`ipfilter` or `git`, will not always work properly in these scenarios.

This middleware will seamlessly and securely read the real IP address from the appropriate header
and replace the proxied IP in the request with the real one. The new IP wil lbe used in Caddy's log
files and with other plugins.

## Syntax
```Caddyfile
realip [cidr] {
    header  name
    from    cidr [cidr... ]
    maxhops hops
    strict
}
```

name is the name of the header containing the actual IP address. Default is  X-Forwarded-For.

cidr is the address range of expected proxy servers. As a security measure, IP headers are only accepted from known proxy servers. Must be a valid CIDR block notation. This may be specified multiple times.

hops is the number of proxy hops allowed. In strict mode, setting this will result in a 403 if it is exceeded. Otherwise it will truncate down to the maximum number of hops and continue processing as otherwise. This can be used in cases where the number of proxies out to the internet will be fixed, but either there are too many CIDR ranges to practically specify or they cannot be known ahead of time.

strict, if specified, will reject requests from unkown proxy IPs with a 403 status. If not specified, it will simply leave the original IP in place.

## CIDR blocks

CIDR is a standard notation for specifying IP ranges. To allow a single IP, use /32 after the ip to specify no mask: 123.222.31.4/32. To allow all IPs (and accept an  X-Forwarded-For header from anybody), use 0.0.0.0/0. Most cloud services should have their ip ranges published in this format somewhere.
#### Example
```Caddyfile
realip {
    from 1.2.3.4/32
    from 2.3.4.5/32
}
```

## Presets

There is a few helpers supplied if you want to use Caddy behind cloud services. Simply specify the caddy snippet below in your caddyfile to activate it using a built-in IP list.

| Provider              | Alias        | Caddyfile Snippet   |
|-----------------------|--------------|---------------------|
| Cloudflare            | `cloudflare` | `realip cloudflare` |
| Google Cloud Platform | `gcp`        | `realip gcp`        |
| Rackspace Cloud       | `rackspace`  | `realip rackspace`  |

Additional presets would be welcome by pull request for other cloud providers.

## Examples

Simple usage to read `X-Forwarded-For` from a few known IPs only:

```Caddyfile
realip {
    from 1.2.3.4/32
    from 2.3.4.5/32
}
```

Simple usage of preset and IP:
```Caddyfile
realip cloudflare 1.2.3.4/32
```
or
```Caddyfile
realip cloudflare {
    from 1.2.3.4/32
}
```
or
```Caddyfile
realip {
    from cloudflare
    from 1.2.3.4/32
}
```
