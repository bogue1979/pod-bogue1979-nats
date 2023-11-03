build:
	CGO_ENABLED=0 go build -o pod-bogue1979-nats -a -ldflags '-extldflags "-static"' .

