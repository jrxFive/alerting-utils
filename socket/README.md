#Socket Check

```
Usage: socket [options]

Options:

  -ip=""            IP of service
  -port=  			Port of IP service, default is 80
  -address=""       Full address of service and port
  -timeout=""       Time duration till connection attempt will be closed
  -input=""         Command to send to socket service, newline appended automatically
  -output=""        Exact expected output from input command


Examples:
	socket -ip 10.0.2.15 -port 8000 -input PING -outpt '+PONG'
	socket --address 10.0.2.15:8000 -input PING -outpt '+PONG'
```