package ui

// interface to inject application into UI to avoid circular dependencies
type Application interface {
	FindReachableIPs(port string) []string
	RunControl(ip string, port string) error
}
