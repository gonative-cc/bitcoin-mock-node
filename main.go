package main

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
)

// simpleAddr implements the net.Addr interface with two struct fields
type simpleAddr struct {
	net, addr string
}

// String returns the address.
//
// This is part of the net.Addr interface.
func (a simpleAddr) String() string {
	return a.addr
}

// Network returns the network.
//
// This is part of the net.Addr interface.
func (a simpleAddr) Network() string {
	return a.net
}

// Ensure simpleAddr implements the net.Addr interface.
var _ net.Addr = simpleAddr{}

// TODO: we should not be needing this function,
// cause our only use case is to run RPC on localhost

// parseListeners determines whether each listen address is IPv4 and IPv6 and
// returns a slice of appropriate net.Addrs to listen on with TCP. It also
// properly detects addresses which apply to "all interfaces" and adds the
// address as both IPv4 and IPv6.
func parseListeners(addrs []string) ([]net.Addr, error) {
	netAddrs := make([]net.Addr, 0, len(addrs)*2)
	for _, addr := range addrs {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			// Shouldn't happen due to already being normalized.
			return nil, err
		}

		// Empty host or host of * on plan9 is both IPv4 and IPv6.
		if host == "" || (host == "*" && runtime.GOOS == "plan9") {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp4", addr: addr})
			netAddrs = append(netAddrs, simpleAddr{net: "tcp6", addr: addr})
			continue
		}

		// Strip IPv6 zone id if present since net.ParseIP does not
		// handle it.
		zoneIndex := strings.LastIndex(host, "%")
		if zoneIndex > 0 {
			host = host[:zoneIndex]
		}

		// Parse the IP.
		ip := net.ParseIP(host)
		if ip == nil {
			return nil, fmt.Errorf("'%s' is not a valid IP address", host)
		}

		// To4 returns nil when the IP is not an IPv4 address, so use
		// this determine the address type.
		if ip.To4() == nil {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp6", addr: addr})
		} else {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp4", addr: addr})
		}
	}
	return netAddrs, nil
}

// setupRPCListeners returns a slice of listeners that are configured for use
// with the RPC server depending on the configuration settings for listen
// addresses and TLS.
func setupRPCListeners() ([]net.Listener, error) {
	// Setup TLS if not disabled.
	listenFunc := net.Listen

	// TODO: check and discuss if TLS certificate needed
	// if !cfg.DisableTLS {
	// 	// Generate the TLS cert and key file if both don't already
	// 	// exist.
	// 	if !fileExists(cfg.RPCKey) && !fileExists(cfg.RPCCert) {
	// 		err := genCertPair(cfg.RPCCert, cfg.RPCKey)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 	}
	// 	keypair, err := tls.LoadX509KeyPair(cfg.RPCCert, cfg.RPCKey)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	tlsConfig := tls.Config{
	// 		Certificates: []tls.Certificate{keypair},
	// 		MinVersion:   tls.VersionTLS12,
	// 	}

	// 	// Change the standard net.Listen function to the tls one.
	// 	listenFunc = func(net string, laddr string) (net.Listener, error) {
	// 		return tls.Listen(net, laddr, &tlsConfig)
	// 	}
	// }

	// TODO: below addr and port should be moved to config file
	addrs, err := net.LookupHost("localhost")
	rpcPort := "8334"
	if err != nil {
		return nil, err
	}
	RPCListeners := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		addr = net.JoinHostPort(addr, rpcPort)
		RPCListeners = append(RPCListeners, addr)
	}

	netAddrs, err := parseListeners(RPCListeners)
	if err != nil {
		return nil, err
	}

	listeners := make([]net.Listener, 0, len(netAddrs))
	for _, addr := range netAddrs {
		listener, err := listenFunc(addr.Network(), addr.String())
		if err != nil {
			// TODO: uncomment when looger is added
			// rpcsLog.Warnf("Can't listen on %s: %v", addr, err)
			continue
		}
		listeners = append(listeners, listener)
	}

	return listeners, nil
}

func main() {
	// set the logging level

	// read the default config file for node (node behaviour cfg)

	// read the input config file (yaml) having all the
	// node data (txns, utxos, blocks etc)

	rpcListeners, err := setupRPCListeners()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load RPC listeners: %v\n", err)
		os.Exit(1)
	}

	// start the rpc server
	rpcServer, err := newRPCServer(&rpcserverConfig{
		Listeners:   rpcListeners,
		StartupTime: s.startupTime,
		// ConnMgr:     &rpcConnManager{&s},
		// SyncMgr:      &rpcSyncMgr{&s, s.syncManager},
		TimeSource:  s.timeSource,
		Chain:       s.chain,
		ChainParams: chainParams,
		// DB:           db,
		TxMemPool: s.txMemPool,
		Generator: blockTemplateGenerator,
		// CPUMiner:     s.cpuMiner,
		TxIndex:      s.txIndex,
		AddrIndex:    s.addrIndex,
		CfIndex:      s.cfIndex,
		FeeEstimator: s.feeEstimator,
	})
}
