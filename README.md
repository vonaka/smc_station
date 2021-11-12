SMC
===

Saturday-Morning Cartoon station.

Usage
-----

Start the station with an appropriate configuration file:

```bash
go run ./smc.go -home ./ -config smc.conf -log smc.log
```

At this point the station is available at `http://localhost:8080`.

Configuration
-------------

```shell
start    8:00AM                  # The time the show starts
each     24h0m0s                 # How frequent it is
duration 3h3m0s                  # The duration of the show
data     /data_dir               # Path to data
static   static                  # Path to static files
ignore   "/data_dir/the x-files" # Files to ignore
```