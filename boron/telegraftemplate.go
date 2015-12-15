package main

//TelegrafPlugin Structure to represent Telegraphite plugin stanza and associated Key/Values
type TelegrafPlugin struct {
	Plugin string
	KV     []string
}

//Generic Template
var TelegrafTemplate = `# Telegraf configuration

[tags]

# Configuration for telegraf agent
[agent]
	# Default data collection interval for all plugins
	interval = "10s"

	# If utc = false, uses local time (utc is highly recommended)
	utc = true

	# Precision of writes, valid values are n, u, ms, s, m, and h
	# note: using second precision greatly helps InfluxDB compression
	precision = "s"

	# run telegraf in debug mode
	debug = false

	# Override default hostname, if empty use os.Hostname()
	hostname = ""


###############################################################################
#                                  OUTPUTS                                    #
###############################################################################

[outputs]

# Configuration for influxdb server to send metrics to
[[outputs.influxdb]]
  urls = ["http://localhost:8086"] # required
  database = "telegraf" # required
  precision = "s"


###############################################################################
#                                  PLUGINS                                    #
###############################################################################

[plugins]
  [[plugins.{{.Plugin}}]]
  {{range .KV}}{{.}}
{{end}}`
