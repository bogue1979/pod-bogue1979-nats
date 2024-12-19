package nats

import (
	"time"

	"github.com/bogue1979/pod-bogue1979-nats/babashka"
	"github.com/nats-io/nats.go"
)

type Sub struct {
	Subject string `json:"subject"`
	Queue   string `json:"queue"`
}

type BBnatsMsg struct {
	Subject      string      `json:"subject"`
	Reply        string      `json:"reply"`
	Header       nats.Header `json:"header"`
	Data         string      `json:"data"`
	Subscription Sub         `json:"subscription"`
}

type BBResponse struct {
	NatsMessage BBnatsMsg `json:"nats_message,omitempty"`
	Error       *string   `json:"error,omitempty"`
}

func toBBResponse(m *nats.Msg) BBResponse {
	return BBResponse{
		BBnatsMsg{
			Subject:      m.Subject,
			Reply:        m.Reply,
			Header:       m.Header,
			Data:         string(m.Data),
			Subscription: Sub{m.Sub.Subject, m.Sub.Queue},
		}, nil}
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
		fail(message, err.Error())
		return
	}

	err = nc.Publish(opts.Subject, []byte(opts.Msg))
	if err != nil {
		fail(message, err.Error())
		return
	}
	babashka.WriteInvokeResponse(message, "ok")
}

func timeoutSeconds(i int) time.Duration {
	if i == 0 {
		i = 1
	}
	return time.Duration(i * int(time.Second))
}

func request(message *babashka.Message, opts Opts) {
	nc, err := natsConnect(opts)
	if err != nil {
		fail(message, err.Error())
		return
	}
	defer nc.Close()

	sub := opts.Subject
	msg := []byte(opts.Msg)
	timeout := timeoutSeconds(opts.TimeoutSeconds)

	resp, err := nc.Request(sub, msg, timeout)
	if err != nil {
		fail(message, err.Error())
		return
	}
	babashka.WriteInvokeResponse(message, toBBResponse(resp))
}

func subscribe(message *babashka.Message, opts Opts) {
	done := make(chan bool)

	nc, err := natsConnect(opts)
	if err != nil {
		fail(message, err.Error())
		return
	}
	defer nc.Close()

	nc.Subscribe(opts.Subject, func(m *nats.Msg) {
		babashka.WriteInvokeResponse(message, toBBResponse(m))
	})
	nc.Flush()
	if err := nc.LastError(); err != nil {
		fail(message, err.Error())
		return
	}
	<-done
}
