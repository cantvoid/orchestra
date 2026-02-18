# orchestra
a VPN orchestrator in a single binary, uses sing-box as core.

### usage
to use orchestra, you will need sing-box. you can get it from [here](https://github.com/SagerNet/sing-box/releases).
```bash
orchestra --link https://fake.com/subscription --singbox-path /usr/bin/sing-box
#automatically starts the best subscription by latency, and fails over when it rotates or stops working.
orchestra -l  https://fake.com/subscription -s /usr/bin/sing-box -wait 10s -timeout 1m
#shorter usage, with more customization
orchestra
#shows all options
```

### compiling from source
to compile from source, install [go](https://go.dev/dl) and [git](https://git-scm.com/install/) first.
```bash
git clone https://github.com/cantvoid/orchestra #clone the git repo
cd orchestra
go build #download dependencies and build orchestra
#now you have an orchestra binary
```
