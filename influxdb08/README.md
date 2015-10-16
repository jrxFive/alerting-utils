#Influxdb Check

```
Usage: influxdb08 [options]
Options:
  -ip=""            IP of InfluxDB service, Required
  -port=            Port of InfluxDB service, default is 8086
  -database=""      Database to Query, Required
  -user=""          Username that has access to the database specified, default root
  -password=""      Password to the username that has access to the InfluxDB database, default root
  -series=""        Full series name of the timeseries you want to check thresholds, Either Series or Custom Required
  -custom=""        Custom Influx Query of the series you want to check thresholds, Either Series or Custom Required
  -delta=           Now() - delta, default 60s
  -count            Compare threshold to count
  -min              Compare threshold to min
  -max              Compare threshold to max
  -mean             Compare threshold to mean
  -mode             Compare threshold to mode
  -median           Compare threshold to median
  -derivative       Compare threshold to derivative
  -sum              Compare threshold to sum
  -stddev           Compare threshold to standard deviation
  -first            Compare threshold to first
  -lessthan         Warning and Critical values will notify if less than
  -warning=         Exits with code 1 if exceeded, Optional
  -critical=        Exits with code 2 if exceeded, Required
Examples:
	influxdb08 -ip="0.0.0.0" -series="servers.consulalerting.cpu.total.idle" -mean -delta=500 -database="db" -critical=200
```