package nats

import (
	"fmt"

	"github.com/bogue1979/pod-bogue1979-nats/babashka"
	"github.com/nats-io/nats.go"
)

type BBKvMsg struct {
	Bucket    string `json:"bucket,omitempty"`
	Key       string `json:"key,omitempty"`
	Value     string `json:"value,omitempty"`
	Revision  uint64 `json:"revision,omitempty"`
	Operation string `json:"operation,omitempty"`
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
	revision, err := bucket.Put(opts.Key, []byte(opts.Value))
	if err != nil {
		sendError(message, err)
		return
	}
	babashka.WriteInvokeResponse(message, BBKvMsg{Revision: revision})
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
	babashka.WriteInvokeResponse(message,
		BBKvMsg{
			Bucket:    entry.Bucket(),
			Key:       entry.Key(),
			Value:     string(entry.Value()),
			Revision:  entry.Revision(),
			Operation: entry.Operation().String(),
		},
	)
}

func kvwatchbucket(message *babashka.Message, opts Opts) {
	if opts.Host == "" {
		fail(message, "please provide :host")
		return
	}

	if opts.Bucket == "" {
		fail(message, "please provide :bucket")
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
	watcher, err := bucket.WatchAll()
	if err != nil {
		sendError(message, err)
		return
	}
	defer watcher.Stop()

	for {
		entry, ok := <-watcher.Updates()
		if !ok {
			break
		}
		if entry != nil {
			babashka.WriteInvokeResponse(message,
				BBKvMsg{
					Bucket:    entry.Bucket(),
					Key:       entry.Key(),
					Value:     string(entry.Value()),
					Revision:  entry.Revision(),
					Operation: entry.Operation().String(),
				},
			)
		}
	}
}
