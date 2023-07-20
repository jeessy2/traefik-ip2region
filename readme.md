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
          version: v0.0.1
  ```

  ```yaml
  http:
    middlewares:
      my-plugins:
        plugin:
          traefik-ip2region:
            dbPath: /opt/plugins-storage/ip2region.xdb
            # headers:
            #  country: "X-Ip2region-City"
            #  province: "X-Ip2region-Province"
            #  city: "X-Ip2region-City"
            #  isp: "X-Ip2region-Isp"
            ban:
              city:
              #  -
              userAgent:
              #  - 
              country:
              #  - 
  ```

- thanks
  - https://github.com/lionsoul2014/ip2region