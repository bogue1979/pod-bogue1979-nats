package nats

import (
	"time"

	"github.com/bogue1979/pod-bogue1979-nats/babashka"
	"github.com/nats-io/nats.go"
)

type Sub struct {
	Subject string `json:"subject,omitempty"`
	Queue   string `json:"queue,omitempty"`
}

type NatsMessage struct {
	Subject      string      `json:"subject,omitempty"`
	Reply        string      `json:"reply,omitempty"`
	Header       nats.Header `json:"header,omitempty"`
	Data         string      `json:"data,omitempty"`
	Subscription Sub         `json:"subscription,omitempty"`
}

type BBResponse struct {
	NatsMessage `json:"nats_message,omitempty"`
	Error       string `json:"error,omitempty"`
}

type PubResult struct {
	Result string `json:"result"`
}

func natsConnect(opts Opts) (nc *nats.Conn, err error) {
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
	if opts.Host == "" {
		fail(message, "please provide :host")
		return
	}

	if opts.Subject == "" {
		fail(message, "please provide :subject")
		return
	}

	if opts.Msg == "" {
		fail(message, "please provide :msg")
		return
	}

	nc, err := natsConnect(opts)
	if err != nil {
		sendError(message, err)
		return
	}

	err = nc.Publish(opts.Subject, []byte(opts.Msg))
	if err != nil {
		sendError(message, err)
		return
	}
	nc.Flush()
	if err := nc.LastError(); err != nil {
		fail(message, err.Error())
		return
	}
	babashka.WriteInvokeResponse(message, PubResult{Result: "ok"})
}

func timeoutSeconds(i int) time.Duration {
	if i == 0 {
		i = 1
	}
	return time.Duration(i * int(time.Second))
}

func request(message *babashka.Message, opts Opts) {
	if opts.Host == "" {
		fail(message, "please provide :host")
		return
	}

	if opts.Subject == "" {
		fail(message, "please provide :subject")
		return
	}

	if opts.Msg == "" {
		fail(message, "please provide :msg")
		return
	}

	nc, err := natsConnect(opts)
	if err != nil {
		sendError(message, err)
		return
	}
	defer nc.Close()

	sub := opts.Subject
	msg := []byte(opts.Msg)
	timeout := timeoutSeconds(opts.TimeoutSeconds)

	resp, err := nc.Request(sub, msg, timeout)
	if err != nil {
		sendError(message, err)
		return
	}
	babashka.WriteInvokeResponse(message,
		NatsMessage{
			Subject:      resp.Subject,
			Reply:        resp.Reply,
			Header:       resp.Header,
			Data:         string(resp.Data),
			Subscription: Sub{resp.Sub.Subject, resp.Sub.Queue},
		})
}

func subscribe(message *babashka.Message, opts Opts) {
	done := make(chan bool)

	nc, err := natsConnect(opts)
	if err != nil {
		sendError(message, err)
		return
	}
	defer nc.Close()

	nc.Subscribe(opts.Subject, func(resp *nats.Msg) {
		babashka.WriteInvokeResponse(message,
			NatsMessage{
				Subject:      resp.Subject,
				Reply:        resp.Reply,
				Header:       resp.Header,
				Data:         string(resp.Data),
				Subscription: Sub{resp.Sub.Subject, resp.Sub.Queue},
			})
	})
	nc.Flush()
	if err := nc.LastError(); err != nil {
		sendError(message, err)
		return
	}
	<-done
}
