- download ip2region.xdb

  ```shell
  wget https://github.com/lionsoul2014/ip2region/raw/master/data/ip2region.xdb
  ```

- config

  ```yaml
    # Static configuration
    experimental:
      plugins:
        example:
          moduleName: github.com/jeessy2/traefik-ip2region
          version: v0.0.7
  ```

  ```yaml
  http:
    middlewares:
      my-plugins:
        plugin:
          traefik-ip2region:
            dbPath: /plugins-local/config/ip2region.xdb
            headers:
              country: "X-Ip2region-Country"
              province: "X-Ip2region-Province"
              city: "X-Ip2region-City"
              isp: "X-Ip2region-Isp"
            ban:
              enabled: false
              country:
              #  - 
              province:
              #  - 
              city:
              #  -
              userAgent:
              #  - 
            whitelist:
              enabled: false
              country:
              #  - 
              province:
              #  - 
              city:
              #  -
              userAgent:
              #  - 

  ```

- k8s
```yaml
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: traefik-ip2region
spec:
  plugin:
    traefik-ip2region:
      dbPath: /plugins-local/config/ip2region.xdb
      ban:
        enabled: false
      whitelist:
        enabled: true
        country:
          - 中国

```

- thanks
  - https://github.com/lionsoul2014/ip2region