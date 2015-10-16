#Boron Check

Boron is a wrapper around [telegraf](https://github.com/influxdb/telegraf), it will create
a custom generated config file and run the telegraf command in STDOUT mode. It will check for
a singular measurement and compare its current value to the set thresholds. This method made
it easy to check for multiple plugin types insteads of having to distribute add dependencies for 
other styled ruby/python checks.

```
Usage: boron [options]
Options:
  -ip=""                   IP of service, default localhost
  -port=                   Port of Service, default 8000
  -protocol                Protocol of Service, default empty ie tcp:// or http://
  -service                 To enable Service parameters in template generation, default false
  -tags                    Tags to determine correct measurement
  -measurement             Single Telegraf timemeasurement to check against
  -plugin                  Telegraf Plugin name
  -plugin-parameters       Parameters for Telegraf Plugin, separator '|'
  -telegraf-location       Absolute path of telegraf binary
  -working-location        Working location to generate temporary plugin files
  -lessthan                Warning and Critical values will notify if less than
  -warning=                Exits with code 1 if exceeded, Optional
  -critical=               Exits with code 2 if exceeded, Required
Examples:
	./boron -plugin mem --working-location . -telegraf-location ./telegraf -measurement 'mem_used' -critical 0
	./boron -plugin cpu --working-location . -telegraf-location ./telegraf -plugin-parameters 'percpu = true|totalcpu = true|drop = ["cpu_time"]' -measurement 'cpu_usage_idle' -critical 0 -tags 'cpu="cpu0"'
	./boron -warning=20971520 -critical=104857600 -measurement 'some.example.com.host.10.0.0.1.port.6379.redis_used_memory' -plugin redis -telegraf-location './telegraf' -service -ip "0.0.0.0" -port 6379 -protocol "tcp://" -plugin-parameters "key = value|key2 = value2"
```