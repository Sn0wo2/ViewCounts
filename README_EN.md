# ViewCounts

> [中文](README.md)

> ViewCounts API is a visitor API counter developed by [**`Golang`**](https://go.dev)

---

## introduce:

You can set an IP to count once, or an IP to access X times to count once, or an IP can only be recorded once for a period of time and supports recording access logs.

---

## `config.yml:`

```
protocol: http
listen_addr: :80
cert_file: server.crt
key_file: server.key
rate_limit: 60
blacklist: [ ]
template_file: index.html
```

---

## Star History

<a href="https://star-history.com/#Sn0wo2/ViewCounts&Date">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=Sn0wo2/ViewCounts&type=Date&theme=dark" />
    <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=Sn0wo2/ViewCounts&type=Date" />
    <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=Sn0wo2/ViewCounts&type=Date" />
  </picture>
</a>

## License

**Please comply with [`GPL 3 Agreement`](LICENSE)**
