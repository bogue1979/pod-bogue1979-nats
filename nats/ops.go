package nats

import (
	"encoding/json"
	"fmt"

	"github.com/bogue1979/pod-bogue1979-nats/babashka"
)

type Opts struct {
	Host           string `json:"host"`
	Nkey           string `json:"nkey"`
	Subject        string `json:"subject"`
	Msg            string `json:"msg"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	Bucket         string `json:"bucket"`
	Key            string `json:"key"`
	Value          string `json:"value"`
}

type Error struct {
	Error string `json:"error"`
}

func sendError(message *babashka.Message, err error) {
	babashka.WriteInvokeResponse(message, Error{
		Error: err.Error(),
	})
}

func fail(message *babashka.Message, err string) {
	babashka.WriteErrorResponse(message, fmt.Errorf("%s", err))
}

func parseOpts(inputList []json.RawMessage) (opts Opts, err error) {
	if len(inputList) != 1 {
		return opts, fmt.Errorf("Wrong number of arguments, expected is only the opts struct")
	}
	if err := json.Unmarshal([]byte(inputList[0]), &opts); err != nil {
		return opts, err
	}
	return opts, nil
}

func ProcessMessage(message *babashka.Message) (any, error) {
	if message.Op == "describe" {
		return &babashka.DescribeResponse{
			Format: "json",
			Namespaces: []babashka.Namespace{
				{Name: "pod.bogue1979.nats",
					Vars: []babashka.Var{
						{Name: "publish"},
						{Name: "request"},
						{Name: "kvput"},
						{Name: "kvget"},
						{Name: "subscribe",
							Code: `(defn subscribe
												([callback]
												 (subscribe callback {}))
												([callback opts]
												 (babashka.pods/invoke
													 "pod.bogue1979.nats"
													 'pod.bogue1979.nats/subscribe*
													 [opts]
													 {:handlers {:success (fn [event]
																									(callback event))
																			 :error   (fn [{:keys [:ex-message :ex-data]}]
																									(binding [*out* *err*]
																										(println "ERROR:" ex-message)))}})
												 nil))`,
						},
						{Name: "kvwatchbucket",
							Code: `(defn kvwatchbucket
												([callback]
												 (kvwatchbucket callback {}))
												([callback opts]
												 (babashka.pods/invoke
													 "pod.bogue1979.nats"
													 'pod.bogue1979.nats/kvwatchbucket*
													 [opts]
													 {:handlers {:success (fn [event]
																									(callback event))
																			 :error   (fn [{:keys [:ex-message :ex-data]}]
																									(binding [*out* *err*]
																										(println "ERROR:" ex-message)))}})
												 nil))`,
						},
					},
				},
			},
		}, nil
	}

	if message.Op == "invoke" {
		inputList, err := message.Arguments()
		if err != nil {
			return nil, err
		}
		opts, err := parseOpts(inputList)
		if err != nil {
			fail(message, err.Error())
		}

		switch message.Var {
		// plain nats
		case "pod.bogue1979.nats/publish":
			publish(message, opts)

		case "pod.bogue1979.nats/subscribe*":
			subscribe(message, opts)

		case "pod.bogue1979.nats/request":
			request(message, opts)

		// kv
		case "pod.bogue1979.nats/kvwatchbucket*":
			kvwatchbucket(message, opts)

		case "pod.bogue1979.nats/kvput":
			kvput(message, opts)

		case "pod.bogue1979.nats/kvget":
			kvget(message, opts)
		}
	}
	return message, nil
}
