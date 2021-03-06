package ipfsproxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	ma "github.com/multiformats/go-multiaddr"

	"github.com/ipfs/ipfs-cluster/config"
)

const configKey = "ipfsproxy"

// Default values for Config.
const (
	DefaultProxyAddr              = "/ip4/127.0.0.1/tcp/9095"
	DefaultNodeAddr               = "/ip4/127.0.0.1/tcp/5001"
	DefaultProxyReadTimeout       = 0
	DefaultProxyReadHeaderTimeout = 5 * time.Second
	DefaultProxyWriteTimeout      = 0
	DefaultProxyIdleTimeout       = 60 * time.Second
)

// Config allows to customize behaviour of IPFSProxy.
// It implements the config.ComponentConfig interface.
type Config struct {
	config.Saver

	// Listen parameters for the IPFS Proxy.
	ProxyAddr ma.Multiaddr

	// Host/Port for the IPFS daemon.
	NodeAddr ma.Multiaddr

	// Maximum duration before timing out reading a full request
	ProxyReadTimeout time.Duration

	// Maximum duration before timing out reading the headers of a request
	ProxyReadHeaderTimeout time.Duration

	// Maximum duration before timing out write of the response
	ProxyWriteTimeout time.Duration

	// Server-side amount of time a Keep-Alive connection will be
	// kept idle before being reused
	ProxyIdleTimeout time.Duration
}

type jsonConfig struct {
	ProxyListenMultiaddress string `json:"proxy_listen_multiaddress"`
	NodeMultiaddress        string `json:"node_multiaddress"`
	ProxyReadTimeout        string `json:"proxy_read_timeout"`
	ProxyReadHeaderTimeout  string `json:"proxy_read_header_timeout"`
	ProxyWriteTimeout       string `json:"proxy_write_timeout"`
	ProxyIdleTimeout        string `json:"proxy_idle_timeout"`
}

// ConfigKey provides a human-friendly identifier for this type of Config.
func (cfg *Config) ConfigKey() string {
	return configKey
}

// Default sets the fields of this Config to sensible default values.
func (cfg *Config) Default() error {
	proxy, _ := ma.NewMultiaddr(DefaultProxyAddr)
	node, _ := ma.NewMultiaddr(DefaultNodeAddr)
	cfg.ProxyAddr = proxy
	cfg.NodeAddr = node
	cfg.ProxyReadTimeout = DefaultProxyReadTimeout
	cfg.ProxyReadHeaderTimeout = DefaultProxyReadHeaderTimeout
	cfg.ProxyWriteTimeout = DefaultProxyWriteTimeout
	cfg.ProxyIdleTimeout = DefaultProxyIdleTimeout

	return nil
}

// Validate checks that the fields of this Config have sensible values,
// at least in appearance.
func (cfg *Config) Validate() error {
	var err error
	if cfg.ProxyAddr == nil {
		err = errors.New("ipfsproxy.proxy_listen_multiaddress not set")
	}
	if cfg.NodeAddr == nil {
		err = errors.New("ipfsproxy.node_multiaddress not set")
	}

	if cfg.ProxyReadTimeout < 0 {
		err = errors.New("ipfsproxy.proxy_read_timeout is invalid")
	}

	if cfg.ProxyReadHeaderTimeout < 0 {
		err = errors.New("ipfsproxy.proxy_read_header_timeout is invalid")
	}

	if cfg.ProxyWriteTimeout < 0 {
		err = errors.New("ipfsproxy.proxy_write_timeout is invalid")
	}

	if cfg.ProxyIdleTimeout < 0 {
		err = errors.New("ipfsproxy.proxy_idle_timeout invalid")
	}

	return err

}

// LoadJSON parses a JSON representation of this Config as generated by ToJSON.
func (cfg *Config) LoadJSON(raw []byte) error {
	jcfg := &jsonConfig{}
	err := json.Unmarshal(raw, jcfg)
	if err != nil {
		logger.Error("Error unmarshaling ipfsproxy config")
		return err
	}

	cfg.Default()

	proxyAddr, err := ma.NewMultiaddr(jcfg.ProxyListenMultiaddress)
	if err != nil {
		return fmt.Errorf("error parsing ipfs_proxy_listen_multiaddress: %s", err)
	}
	nodeAddr, err := ma.NewMultiaddr(jcfg.NodeMultiaddress)
	if err != nil {
		return fmt.Errorf("error parsing ipfs_node_multiaddress: %s", err)
	}

	cfg.ProxyAddr = proxyAddr
	cfg.NodeAddr = nodeAddr

	err = config.ParseDurations(
		"ipfsproxy",
		&config.DurationOpt{Duration: jcfg.ProxyReadTimeout, Dst: &cfg.ProxyReadTimeout, Name: "proxy_read_timeout"},
		&config.DurationOpt{Duration: jcfg.ProxyReadHeaderTimeout, Dst: &cfg.ProxyReadHeaderTimeout, Name: "proxy_read_header_timeout"},
		&config.DurationOpt{Duration: jcfg.ProxyWriteTimeout, Dst: &cfg.ProxyWriteTimeout, Name: "proxy_write_timeout"},
		&config.DurationOpt{Duration: jcfg.ProxyIdleTimeout, Dst: &cfg.ProxyIdleTimeout, Name: "proxy_idle_timeout"},
	)
	if err != nil {
		return err
	}

	return cfg.Validate()
}

// ToJSON generates a human-friendly JSON representation of this Config.
func (cfg *Config) ToJSON() (raw []byte, err error) {
	// Multiaddress String() may panic
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()

	jcfg := &jsonConfig{}

	// Set all configuration fields
	jcfg.ProxyListenMultiaddress = cfg.ProxyAddr.String()
	jcfg.NodeMultiaddress = cfg.NodeAddr.String()
	jcfg.ProxyReadTimeout = cfg.ProxyReadTimeout.String()
	jcfg.ProxyReadHeaderTimeout = cfg.ProxyReadHeaderTimeout.String()
	jcfg.ProxyWriteTimeout = cfg.ProxyWriteTimeout.String()
	jcfg.ProxyIdleTimeout = cfg.ProxyIdleTimeout.String()

	raw, err = config.DefaultJSONMarshal(jcfg)
	return
}
