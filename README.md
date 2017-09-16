# lsrv

lsrv is a tool that helps manage multiple services that use the same protocol and one machine.
It was born out of burnout for trying to remember port numbers for each thing running on my machine.
Instead, each service can be assigned the default port for the protocol that is intended to be used.
For example, instead of trying to remember which port grafana runs on, and which on prometheus runs on,
we just use port 80 for both, and reference each service by name: `grafana.svc` and `prometheus.svc`.

This tool uses iptables to assign each service an IP address, and then DNAT to remap to localhost
and the correct port.

## Usage
We can add a service as follows:

```
# ./bin/lsrv add grafana 3000 80
```

After running this command, `http://grafana.svc` will forward to `127.0.0.1:3000`. The command modifies
 `/etc/hosts`, however this may change to use an nsswitch module in the future. The iptables
rules are added to the `nat` table under the `LSRV` chain. There will also be a rule to jump to that
chain from the `OUTPUT` chain.

You can ask the cli tool for the IP address:

```
# ./bin/lsrv resolve grafana
```

If we no longer wanted grafana to be mapped:

```
# ./bin/lsrv rm grafana
```

Changes may not persist across a reboot. You should restore the previous state after a reboot:
```
# ./bin/lsrv restore
```

You can cleanup the hosts file and iptables with the following command:
```
# ./bin/lsrv cleanup
```

## Todo
There are some limitations I hope to fix:

- Only one entry per service name is allowed
- Currently, only a start ip address is provided. A proper config should
  allow ip ranges and it's probably worth checking we remain in those bounds
- There is a lack for configuration. A lot of things like the TLD, start ip,
  state file location, etc are hard coded
- The CLI only allows creating ip addresses that forward to 127.0.0.1. This is
  purely a limitation of the CLI interface. The server doesn't care
