# btest_loadbalancer

> prepare
- macos:
    - `sudo ifconfig utun19 10.1.0.10 10.1.0.1 netmask 255.255.255.0 up`
- unix:
    - `sudo ip addr add 10.1.0.10/24 dev utun19 && sudo ip link set dev utun19 up`
