# DNSter

Aggressively parallel DNS utility, written in Go.

**WARNING: DNSter is still under development and is not ready for use.**

## Introduction

Many DNS utilities, like `systemd-resolved` support multiple DNS servers, but
only try secondary servers when primary servers don't respond.

In some scenarios (e.g. when switching between VPNs which have
mutually-exclusive DNS servers), this causes sluggishness, or in some cases,
breaks completely.

DNSter solves this by being as greedy as possible - by contacting every DNS
server in parallel, and returning the first non-empty successful answer.

This consumes significantly more network than ordinary single-DNS approach.

***TODO: DNSter should redeem itself by caching answers***

## Usage

```
Usage of dnster:
  -addr string
        address and port to listen on (default "127.0.0.53:53")
  -conf string
        file containing upstream DNS servers to use (default "dnster.conf")
```

## Install

To replace other DNS services, e.g. `systemd-resolved`, first make sure they
are stopped and disabled:

```bash
sudo systemctl stop systemd-resolved
sudo systemctl disable systemd-resolved
```

Then change the content of `/etc/resolv.conf` to the following:

```
nameserver 127.0.0.53
```

Then clone this repository and run `make && sudo make install`:

```bash
git clone https://github.com/paskozdilar/dnster.git
cd dnster
```

Then edit the `dnster.conf` file inside the repository, adding or removing DNS
servers as preferred.

Then run the following commands:

```
make
sudo make install
```

The DNSter will be running and used as default DNS service.

To change list of DNS servers, edit the `/etc/dnster.conf` file and restart
DNSter:

```
sudo systemctl restart dnster
```

## Uninstall

To uninstall DNSter, simply run `sudo make uninstall` inside the repository.

If necessary, start the previous DNS service:

```
sudo systemctl enable systemd-resolved
sudo systemctl start systemd-resolved
```
