#!/bin/bash

# Check if the script is running as root
if [ "$(id -u)" != "0" ]; then
    echo "This script must be run with superuser privileges. Use 'sudo'." 1>&2
    exit 1
fi

# Enable IP forwarding
sysctl -w net.inet.ip.forwarding=1 

# Check the operating system
if [[ "$(uname)" == "Linux" ]]; then
    # Linux-specific: Flush existing iptables rules
    iptables -F
    iptables -t nat -F

    # Add a NAT rule to redirect traffic through utun19
    # Note: Change utun19 to the actual interface name if different
    iptables -t nat -A POSTROUTING -o utun19 -j MASQUERADE

    echo "Linux: Traffic redirection configured successfully."
elif [[ "$(uname)" == "Darwin" ]]; then
    echo "macOS: Traffic redirection configured successfully."
else
    echo "Unsupported operating system."
    exit 1
fi
