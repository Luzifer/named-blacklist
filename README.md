# Luzifer / named-blacklist

This repo contains a DNS blacklist generator to be used in combination with [DNS Response Policy Zones](https://www.dnsrpz.info/) in BIND 9.8+.

## Usage

For full usage inside BIND see [BIND ARM](https://downloads.isc.org/isc/bind9/9.14.3/doc/arm/Bv9ARM.ch05.html#options).

Quoted from the ARM:

> For example, you might use this option statement

```
response-policy { zone "badlist"; };
```

> and this zone statement

```
zone "badlist" { 
  type master; 
  file "master/badlist"; 
  allow-query { none; }; 
};
```

Then you can generate the `master/badlist` file using `named-blacklist`:

```console
# named-blacklist --config config.sample.yaml | tee master/badlist
$TTL 1H

@ SOA LOCALHOST. dns-master.localhost. (1 1h 15m 30d 2h)
  NS  LOCALHOST.

; Blacklist entries
0.nextyourcontent.com CNAME . ; From: "Dan Pollock - someonewhocares"
0.r.msn.com CNAME . ; From: "add.Risk"
000.0x1f4b0.com CNAME . ; From: "CoinBlocker"
000.gaysexe.free.fr CNAME . ; From: "Mitchell Krog's - Badd Boyz Hosts"
[...]
```

## Provider thresholds

Each provider can define an optional `min_matches` value. It defaults to `1`,
which keeps the previous behavior.

For blacklist providers, a domain is included when the number of blacklist
providers containing that domain is greater than or equal to the lowest
`min_matches` value among those matching providers. This allows trusted feeds
to stay at `1` while noisier feeds can require `2` or `3` matching providers.

Whitelist providers also accept `min_matches`, but the value is ignored:
whitelisted domains are always removed from the final blacklist.
