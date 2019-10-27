
## Overview

```text
$ netstat -ano  | grep EST | grep 8081 
tcp        0      0 127.0.0.1:56702         127.0.0.1:8081          ESTABLISHED keepalive (54.41/0/0)
tcp6       0      0 127.0.0.1:8081          127.0.0.1:56702         ESTABLISHED keepalive (7.16/0/0)
```

- TCP keepalive是双向的。Client/Server的Socket都会有各自的timer。
- Microsoft SDN的timeout是`4min = 240s`
- `SO_KEEPALIVE`
- redis server的default tcp keepalive is 300s
- App-Level vs TCP-level keepalive parameters

## Linux OS

```text
/proc/sys/net/ipv4/tcp_keepalive_time    当keepalive起用的时候，TCP发送keepalive消息的频度。缺省是2小时。
/proc/sys/net/ipv4/tcp_keepalive_intvl   当探测没有确认时，重新发送探测的频度。缺省是75秒。
/proc/sys/net/ipv4/tcp_keepalive_probes  在认定连接失效之前，发送多少个TCP的keepalive探测包。缺省值是9。这个值乘以tcp_keepalive_intvl之后决定了，一个连接发送了keepalive之后可以有多少时间没有回应
```

```text
# vi /etc/sysctl.conf
net.ipv4.tcp_keepalive_time = 60
net.ipv4.tcp_keepalive_intvl = 10
net.ipv4.tcp_keepalive_probes = 6
# sysctl -p
```

## SSH的timeout

```text
ssh -o ServerAliveInterval=5 -o ServerAliveCountMax=1 $HOST
```

## TODO

- [ ] Scripts for checking conn for non-underlay address(`10.*.*.*`)
