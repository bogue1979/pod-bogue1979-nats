package nats

import (
	"fmt"

	"github.com/bogue1979/pod-bogue1979-nats/babashka"
	"github.com/nats-io/nats.go"
)

type BBKvWatchMsg struct {
	Bucket    string `json:"bucket"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Revision  uint64 `json:"revision"`
	Operation string `json:"operation"`
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

func kvput(message *babashka.Message, opts Opts) {
	if opts.Bucket == "" {
		sendError(message, fmt.Errorf("%s", "please provide :bucket"))
		return
	}
	if opts.Value == "" {
		sendError(message, fmt.Errorf("%s", "please provide :value"))
		return
	}
	if opts.Key == "" {
		sendError(message, fmt.Errorf("%s", "please provide :key"))
		return
	}
	nc, err := natsConnect(opts)
	if err != nil {
		sendError(message, err)
		return
	}
	defer nc.Close()
	js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		sendError(message, err)
		return
	}
	bucket, err := js.KeyValue(opts.Bucket)
	if err != nil {
		sendError(message, err)
		return
	}
	_, err = bucket.Put(opts.Key, []byte(opts.Value))
	if err != nil {
		sendError(message, err)
		return
	}
	babashka.WriteInvokeResponse(message, "ok")
}

func kvget(message *babashka.Message, opts Opts) {
	if opts.Bucket == "" {
		sendError(message, fmt.Errorf("%s", "please provide :bucket"))
		return
	}
	if opts.Key == "" {
		sendError(message, fmt.Errorf("%s", "please provide :key"))
		return
	}
	nc, err := natsConnect(opts)
	if err != nil {
		sendError(message, err)
		return
	}
	defer nc.Close()
	js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		sendError(message, err)
		return
	}
	bucket, err := js.KeyValue(opts.Bucket)
	if err != nil {
		sendError(message, err)
		return
	}
	entry, err := bucket.Get(opts.Key)
	if err != nil {
		sendError(message, err)
		return
	}
	babashka.WriteInvokeResponse(message, parseKVentry(entry))
}

func kvwatchbucket(message *babashka.Message, opts Opts) {
	nc, err := natsConnect(opts)
	if err != nil {
		sendError(message, err)
		return
	}
	defer nc.Close()
	js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		sendError(message, err)
		return
	}
	bucket, err := js.KeyValue(opts.Bucket)
	if err != nil {
		sendError(message, err)
		return
	}
	watcher, err := bucket.WatchAll()
	if err != nil {
		sendError(message, err)
		return
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
