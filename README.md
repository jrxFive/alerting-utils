#Alerting Utilities

A set of individual golang CLI tools. Follows the practice Consul checks, but can be modified for Nagios/Sensu.

##Consul style [Consul Checks](https://consul.io/docs/agent/checks.html)
* Exit code 0 - Check is passing
* Exit code 1 - Check is warning
* Any other code - Check is failing

##Sensu stlye [Sensu Checks](https://sensuapp.org/docs/latest/checks)
* 0 for OK
* 1 for WARNING
* 2 for CRITICAL
* 3 or greater to indicate UNKNOWN or CUSTOM