# remote-write_stale_nans

This tries to aid with finding a bug in Cortex chunks' ruler

It remote writes series to url

node1 stale nans with a 0.01 probability 
node2 stale nans with a 0.1 probability 
node3 no stale nans

```
$ ./remote-write-stale-nans --help
Usage of ./remote-write-stale-nans:
  -send-interval duration
        interval how often series is remote written (default 10s)
  -url string
        remote write url (default "http://cortex:9009/api/v1/push")
```
