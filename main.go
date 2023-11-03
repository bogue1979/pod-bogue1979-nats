package main

import (
	"github.com/bogue1979/pod-bogue1979-nats/babashka"
	"github.com/bogue1979/pod-bogue1979-nats/nats"
)

func main() {

	for {
		message, err := babashka.ReadMessage()
		if err != nil {
			babashka.WriteErrorResponse(message, err)
			continue
		}
		res, err := nats.ProcessMessage(message)
		if err != nil {
			babashka.WriteErrorResponse(message, err)
			continue
		}

		describeRes, ok := res.(*babashka.DescribeResponse)
		if ok {
			babashka.WriteDescribeResponse(describeRes)
			continue
		}
		babashka.WriteInvokeResponse(message, res)
	}
}
