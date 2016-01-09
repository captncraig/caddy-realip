Middleware for restoring real ip information when running caddy behind a proxy. Will allow other middlewares to simply use `r.RemoteAddr` instead of decoding `X-Forwarded-For` themselves.

Analogous to nginx's [realip_module](http://nginx.org/en/docs/http/ngx_http_realip_module.html)

Checks whitelist of authorized proxy servers so we don't arbitrarily trust headers from anybody.

Example config (for servers behind cloudflare)

```
:80 {
    root /var/www/home
    realip {
        from 103.21.244.0/22
        from 103.22.200.0/22
        from 103.31.4.0/22
        from 104.16.0.0/12
        from 108.162.192.0/18
        from 141.101.64.0/18
        from 162.158.0.0/15
        from 172.64.0.0/13
        from 173.245.48.0/20
        from 188.114.96.0/20
        from 190.93.240.0/20
        from 197.234.240.0/22
        from 198.41.128.0/17
        from 199.27.128.0/21
        from 2400:cb00::/32
        from 2405:8100::/32
        from 2405:b500::/32
        from 2606:4700::/32
        from 2803:f800::/32
        header X-Forwarded-For
    }
}
```