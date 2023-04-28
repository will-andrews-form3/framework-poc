## Example

This example uses the framework to create a NATs client and NATs subscription (which depends on the client). It then starts the service and shuts down after 10 seconds. Once the service has started, in a different go routine, some NATs messages are sent to the configured subject.

To start, run `docker-compose up` and then `go run main.go`