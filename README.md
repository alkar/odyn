# odyn
odyn is a dynamic ip address updater for the new age.

It supports a number of public IP address provider web services and can handle AWS Route53 DNS zones.

# help
For help with using the command line tool, please download the binary from the releases and run `odyn --help`.

You can build from source by using the provided `Dockerfile`:

```
$ docker build -t odyn .
$ docker run --rm -it odyn --help
```

For documentation on how to use this package, please see the [docs](https://godoc.org/github.com/alkar/odyn).

# contributing
Pull requests that extend the supported services or introduce other improvements are welcome. Feel free to open an issue and discuss if you're unsure.
