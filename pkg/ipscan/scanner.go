package ipscan

type Scanner interface {
	Scan(subnet string) ([]Device, error)
	Ping(ip string) error
	GetLocalSubnet() string
}
