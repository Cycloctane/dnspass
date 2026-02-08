package config

import (
	"github.com/BurntSushi/toml"
	"github.com/Cycloctane/dnspass/internal/records"
)

type Config struct {
	Server  ServerConfig  `toml:"server"`
	DNS     DNSConfig     `toml:"dns"`
	Records []RecordEntry `toml:"records"`
}

type ServerConfig struct {
	DNSListen    string `toml:"dns_listen"`
	DNSProtocol  string `toml:"dns_protocol"`
	ProxyListen  string `toml:"proxy_listen"`
	ProxyBlockIP bool   `toml:"proxy_block_ip_connections"`
	WebListen    string `toml:"web_listen"`
	WebUserName  string `toml:"web_username"`
	WebPassWord  string `toml:"web_password"`
}

type DNSConfig struct {
	Upstream []string `toml:"upstream"`
}

type RecordEntry struct {
	Name  string `toml:"name"`
	Type  string `toml:"type"`
	Value string `toml:"value"`
	TTL   uint32 `toml:"ttl"`
}

func Load(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) ApplyRecords() {
	items := make([]records.Record, 0, len(c.Records))
	for _, r := range c.Records {
		items = append(items, records.Record{
			Name:  r.Name,
			Type:  records.DNSType(r.Type),
			Value: r.Value,
			TTL:   r.TTL,
		})
	}
	records.SetAll(items)
}
