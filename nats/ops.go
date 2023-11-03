package nats

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bogue1979/pod-bogue1979-nats/babashka"
	"github.com/nats-io/nats.go"
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

type BBResponse struct {
	NatsMessage BBnatsMsg `json:"nats_message,omitempty"`
	Error       *string   `json:"error,omitempty"`
}

type BBnatsMsg struct {
	Subject      string      `json:"subject"`
	Reply        string      `json:"reply"`
	Header       nats.Header `json:"header"`
	Data         string      `json:"data"`
	Subscription Sub         `json:"subscription"`
}

type BBKvWatchMsg struct {
	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Revision  uint64 `json:"revision"`
	Operation string `json:"operation"`
}

type Sub struct {
	Subject string `json:"subject"`
	Queue   string `json:"queue"`
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
						{Name: "subscribe",
							Code: `
(defn subscribe
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
   nil))
`,
						},
						{Name: "kvwatchbucket",
							Code: `
(defn kvwatchbucket
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
   nil))
`,
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
			babashka.WriteErrorResponse(message, err)
		}

		switch message.Var {

		case "pod.bogue1979.nats/kvwatchbucket*":
			kvwatchbucket(message, opts)

		case "pod.bogue1979.nats/kvput":
			kvput(message, opts)

		case "pod.bogue1979.nats/subscribe*":
			subscribe(message, opts)

		case "pod.bogue1979.nats/publish":
			publish(message, opts)

		case "pod.bogue1979.nats/request":
			request(message, opts)

		}
	}

	return message, nil
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

func connect(opts Opts) (nc *nats.Conn, err error) {

	if opts.Nkey != "" {
		auth, err := AuthFromSeed(opts.Nkey)
		if err != nil {
			return nil, err
		}
		return nats.Connect(opts.Host, auth)
	}
	return nats.Connect(opts.Host)
}

func publish(message *babashka.Message, opts Opts) {

	if opts.Msg == "" {
		babashka.WriteErrorResponse(message, fmt.Errorf("can not send empty message"))
	}

	nc, err := connect(opts)
	if err != nil {
		babashka.WriteErrorResponse(message, err)
	}

	err = nc.Publish(opts.Subject, []byte(opts.Msg))
	if err != nil {
		babashka.WriteErrorResponse(message, err)
	}
	babashka.WriteInvokeResponse(message, "ok")
}

func kvput(message *babashka.Message, opts Opts) {
	if opts.Msg == "" {
		babashka.WriteErrorResponse(message, fmt.Errorf("can not send empty message"))
	}
	nc, err := connect(opts)
	if err != nil {
		babashka.WriteErrorResponse(message, err)
	}
	defer nc.Close()
	js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		babashka.WriteErrorResponse(message, err)
	}
	bucket, err := js.KeyValue(opts.Bucket)
	if err != nil {
		babashka.WriteErrorResponse(message, err)
	}
	_, err = bucket.Put(opts.Key, []byte(opts.Value))
	if err != nil {
		babashka.WriteErrorResponse(message, err)
	}
	babashka.WriteInvokeResponse(message, "ok")
}

func parseKVentry(in nats.KeyValueEntry) BBKvWatchMsg {
	return BBKvWatchMsg{
		Bucket:    in.Bucket(),
		Key:       in.Key(),
		Value:     string(in.Value()),
		Revision:  in.Revision(),
		Operation: in.Operation().String(),
	}
}

func dataToString(m *nats.Msg) BBResponse {
	return BBResponse{
		BBnatsMsg{
			Subject:      m.Subject,
			Reply:        m.Reply,
			Header:       m.Header,
			Data:         string(m.Data),
			Subscription: Sub{m.Sub.Subject, m.Sub.Queue},
		}, nil}
}

func timeoutSeconds(i int) time.Duration {
	if i == 0 {
		i = 1
	}
	return time.Duration(i * int(time.Second))
}

func request(message *babashka.Message, opts Opts) {
	nc, err := connect(opts)
	if err != nil {
		babashka.WriteErrorResponse(message, err)
	}
	defer nc.Close()

	sub := opts.Subject
	msg := []byte(opts.Msg)
	timeout := timeoutSeconds(opts.TimeoutSeconds)

	resp, err := nc.Request(sub, msg, timeout)
	if err != nil {
		babashka.WriteErrorResponse(message, err)
	}
	babashka.WriteInvokeResponse(message, dataToString(resp))
}

func kvwatchbucket(message *babashka.Message, opts Opts) {
	nc, err := connect(opts)
	if err != nil {
		babashka.WriteErrorResponse(message, err)
	}
	defer nc.Close()
	js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		babashka.WriteErrorResponse(message, err)
	}
	bucket, err := js.KeyValue(opts.Bucket)
	if err != nil {
		babashka.WriteErrorResponse(message, err)
	}
	watcher, err := bucket.WatchAll()
	if err != nil {
		babashka.WriteErrorResponse(message, err)
	}
	defer watcher.Stop()

	for {
		current, ok := <-watcher.Updates()
		if !ok {
			break
		}
		if current != nil {
			babashka.WriteInvokeResponse(message, parseKVentry(current))
		}
	}
}

func subscribe(message *babashka.Message, opts Opts) {
	done := make(chan bool)

	nc, err := connect(opts)
	if err != nil {
		babashka.WriteErrorResponse(message, err)
	}
	defer nc.Close()

	nc.Subscribe(opts.Subject, func(m *nats.Msg) {
		babashka.WriteInvokeResponse(message, dataToString(m))
	})
	nc.Flush()
	if err := nc.LastError(); err != nil {
		babashka.WriteErrorResponse(message, err)
	}
	<-done
}
