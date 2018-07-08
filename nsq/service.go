/*
 * Revision History:
 *     Initial: 2018/06/25        Wang RiYu
 */

package nsq

import (
	"io/ioutil"

	"github.com/nsqio/nsq/nsqadmin"
	"github.com/nsqio/nsq/nsqd"
	"github.com/nsqio/nsq/nsqlookupd"
)

// LookupdOptions https://nsq.io/components/nsqlookupd.html
// nsqlookupd.NewOptions() return default Option
type LookupdOptions = nsqlookupd.Options

// NsqdOptions https://nsq.io/components/nsqd.html
// nsqd.NewOptions() return default Option
type NsqdOptions = nsqd.Options

// NsqadminOptions https://nsq.io/components/nsqadmin.html
// nsqadmin.NewOptions() return default Option
type NsqadminOptions = nsqadmin.Options

// NSQLookupd is the daemon that manages topology information.
// Clients query nsqlookupd to discover nsqd producers for a specific topic
// and nsqd nodes broadcasts topic and channel information.
//
// There are two interfaces:
// A TCP interface which is used by nsqd for broadcasts,
// and an HTTP interface for clients to perform discovery and administrative actions.
type NSQLookupd = nsqlookupd.NSQLookupd

// NSQD is the daemon that receives, queues, and delivers messages to clients.
// It can be run standalone but is normally configured in a cluster with nsqlookupd instance(s) (in which case it will announce topics and channels for discovery).
// It listens on two TCP ports, one for clients and another for the HTTP API.
// It can optionally listen on a third port for HTTPS.
type NSQD = nsqd.NSQD

// NSQAdmin is a Web UI to view aggregated cluster stats in realtime and perform various administrative tasks.
type NSQAdmin = nsqadmin.NSQAdmin

// StartNsqlookupd start a nsqlookupd service
func StartNsqlookupd(opts *LookupdOptions) (*NSQLookupd, error) {
	if opts == nil {
		opts = nsqlookupd.NewOptions()
		opts.LogLevel = "warn" // github.com/nsqio/nsq/internal/lg/lg.go
	}

	lookupd := nsqlookupd.New(opts)

	if err := lookupd.Main(); err != nil {
		return nil, err
	}

	return lookupd, nil
}

// StartNsqd start a nsqd service
func StartNsqd(opts *NsqdOptions) (*NSQD, error) {
	if opts == nil {
		opts = nsqd.NewOptions()
	}

	if opts.DataPath == "" {
		tmpDir, err := ioutil.TempDir("", "nsq-test-")
		if err != nil {
			return nil, err
		}
		opts.DataPath = tmpDir
	}

	nsqd := nsqd.New(opts)
	nsqd.Main()

	return nsqd, nil
}

// StartNsqAdmin start a nsqadmin service
func StartNsqAdmin(opts *NsqadminOptions) (*NSQAdmin, error) {
	if opts == nil {
		opts = nsqadmin.NewOptions()
		opts.LogLevel = "warn"
	}

	nsqAdmin := nsqadmin.New(opts)
	nsqAdmin.Main()

	return nsqAdmin, nil
}

// NewLookupdOptions ...
func NewLookupdOptions() *LookupdOptions {
	return nsqlookupd.NewOptions()
}

// NewNsqdOptions ...
func NewNsqdOptions() *NsqdOptions {
	return nsqd.NewOptions()
}

// NewNsqadminOptions ...
func NewNsqadminOptions() *NsqadminOptions {
	return nsqadmin.NewOptions()
}
