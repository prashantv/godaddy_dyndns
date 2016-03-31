# GoDaddy Dynamic DNS
A simple dynamic DNS updater for GoDaddy

## Installing

[Install Go](https://golang.org/doc/install) and run:
```
go get -v -u -f github.com/prashantv/godaddy_dyndns
```

This will install the `godaddy_dyndns` binary to `$GOPATH/bin/godaddy_dns`.

## Configuring

1. Create an `A` record for a subdomain using the [GoDaddy DNS manager](https://dcc.godaddy.com/manage/).
   Record the subdomain you created and the domain under which the subdomain was created.
3. Get a **production** key from [the Keys page](https://developer.godaddy.com/keys).
2. Create a `secrets.json` file that contains the key contents in the following format:
```json
{
  "apiKey": "--INSERT-API-KEY--",
  "apiSecret": "--INSERT-API-SECRET--"
}
```

## Running
```
Usage of godaddy_dyndns:
  -root-domain string
    	The root GoDaddy domain (default "domain.com")
  -secrets-file string
    	Path to a file containing the Godaddy API key and secret (default "secrets.json")
  -sub-domain string
    	The subdomain to update (default "sub")
```

A simple command line is:
```
godaddy_dyndns --secrets-file secrets.json  --root-domain example.com --sub-domain home
```

This will update `home.example.com` to the external IP that is detected according to [MyExternalIP](http://myexternalip.com/).

You can set up this job to run on a schedule (e.g. via Cron).
