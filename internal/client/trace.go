package client

import (
	"fmt"
	"net"
	"net/http/httptrace"
	"time"
)

type Debug struct {
	Address string
	DNS     struct {
		Start   time.Time
		End     time.Time
		Host    string
		Address []net.IPAddr
		Error   error
	}
	Dial struct {
		Start time.Time
		End   time.Time
	}
	Connection struct {
		Start time.Time
		End   time.Time
	}
	WroteHeaders struct {
		Time time.Time
	}
	GotFirstResponseByte struct {
		Time time.Time
	}
	Request struct {
		Start time.Time
		End   time.Time
	}
}

func (d *Debug) String() string {
	var out = `
Address: %s 
DNS Duration: %s 
Request Duration: %s
`

	return fmt.Sprintf(out,
		d.Address,
		d.DNS.End.Sub(d.DNS.Start).String(),
		d.Request.End.Sub(d.Request.Start).String(),
	)
}

func trace() (*httptrace.ClientTrace, *Debug) {
	d := &Debug{}

	t := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			d.Address = hostPort
			d.Connection.Start = time.Now().UTC()
		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			d.Request.Start = time.Now().UTC()
		},
		GotFirstResponseByte: func() {
			d.GotFirstResponseByte.Time = time.Now().UTC()
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			d.DNS.Start = time.Now().UTC()
			d.DNS.Host = info.Host
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			d.DNS.End = time.Now().UTC()
			d.DNS.Address = info.Addrs
			d.DNS.Error = info.Err
		},
		ConnectStart: func(network, addr string) {
			d.Dial.Start = time.Now().UTC()
		},
		ConnectDone: func(network, addr string, err error) {
			d.Dial.End = time.Now().UTC()
		},
		WroteHeaders: func() {
			d.WroteHeaders.Time = time.Now().UTC()
		},
		WroteRequest: func(wr httptrace.WroteRequestInfo) {
			d.Request.End = time.Now().UTC()
		},
	}

	return t, d
}
