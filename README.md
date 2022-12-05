# World Cup Dashboard

This is a simple FIFA World Cup "playoff" dashboard that can be used to show live game scores in browser kiosk mode. 

It's been primarily made to run on [Google Cloud](https://cloud.google.com/)'s [AppEngine](https://cloud.google.com/appengine), but
can run standalone, and probably other clouds.

## Issues and Feature Requests

Please, [submit an issue](https://github.com/icemarkom/worldcup/issues/new).

## Data Sources

Data source is [World Cup JSON](https://worldcupjson.net/). There is *minimal* cosmetic data processing, and refresh rate of the
game pages is set to a reasonable value.

## FAQ

### Wow, this is great - is this live somewhere on the Internet?

Yes, but please run your own instance instead.

### I tried running this but I get the error below:

```shell
$ go run main/main.go
worldcup.go:4:2: package embed is not in GOROOT (/usr/lib/go-1.15/src/embed)
```

You need at least [Go 1.16](https://go.dev/doc/go1.16#library-embed).

### Why make this when there are so many other options?

I had specific requirements that I wanted to meet, and no other option provided those. Also, because I wanted to.
