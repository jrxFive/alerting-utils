#HTTP Check

```
Usage: http [options]

Options:

  -ip=""            IP of service
  -port=  			Port of IP service, default is 80
  -endpoint=""      Endpoint of service if not just the domain
  -address=""       Full 'www' address, port, and endpoint
  -timeout=""       Time duration till connection attempt will be closed


Examples:
	http -ip 10.0.2.15 -port 8000 -endpoint /admin
	http --ip=10.0.2.15 --port=8000 
	http -address=www.google.com
	http --address=www.yahoo.com
```