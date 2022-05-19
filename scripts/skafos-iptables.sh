#!/bin/bash

# Args:
#  1. Container ID

redirect_chain_name=SKAFOS_REDIRECT

# Get container's PID.
pid=$(docker inspect $1 -f '{{.State.Pid}}')
if [ $? -ne 0 ]
    then exit $?
fi

# Change iptables of the container's network namespace.
nsenter -t $pid -n bash <<EOF
iptables -t nat -N $redirect_chain_name
iptables -t nat -A OUTPUT -p tcp -m owner --uid-owner 1234 -j ACCEPT
iptables -t nat -A OUTPUT -p tcp -j $redirect_chain_name
iptables -t nat -A $redirect_chain_name -p tcp -j REDIRECT --to-ports 16000
exit
EOF
