package config

import (
	"github.com/InRaining/NoDelay/common/set"
	outbound2 "github.com/InRaining/NoDelay/outbound"
)

type configMain struct {
	Services []*ConfigProxyService
	Configuration *Configure
	TrafficLimiter *TrafficLimiterConfig
	Lists    map[string]set.StringSet
}

type ConfigProxyService struct {
	Name          string
	TargetAddress string
	TargetPort    uint16
	Listen        uint16
	Flow          string

	IPAccess      access                   `json:",omitempty"`
	Minecraft     minecraft                `json:",omitempty"`
	TLSSniffing   tlsSniffing              `json:",omitempty"`
	SocketOptions *outbound2.SocketOptions `json:",omitempty"`
	Outbound      outbound                 `json:",omitempty"`
}

type access struct {
	Mode     string   // 'accept' or 'deny' or empty
	ListTags []string `json:",omitempty"`
}

type minecraft struct {
	EnableHostnameRewrite bool
	RewrittenHostname     string `json:",omitempty"`

	EnableHostnameAccess bool
	HostnameAccess       string `json:",omitempty"`

	OnlineCount onlineCount

	IgnoreFMLSuffix bool `json:",omitempty"`

	NameAccess access `json:",omitempty"`

	EnableAnyDest   bool          `json:",omitempty"`
	AnyDestSettings configAnyDest `json:",omitempty"`

	PingMode        string
	MotdFavicon     string
	MotdDescription string
}

type onlineCount struct {
	Max            int
	Online         int32
	EnableMaxLimit bool
	Sample         any `json:",omitempty"`
}

type Sample struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type configAnyDest struct {
	WildcardRootDomainName string `json:",omitempty"`
}

type tlsSniffing struct {
	RejectNonTLS     bool
	RejectIfNonMatch bool     `json:",omitempty"`
	SNIAllowListTags []string `json:",omitempty"`
}

type outbound struct {
	Type    string
	Network string `json:",omitempty"`
	Address string `json:",omitempty"`
}

type Configure struct {
	ListAPI        string
	Header         string
	ContactName    string
	ContactLink    string
	WebLogPort     uint16 `json:",omitempty"`
}

type TrafficLimiterConfig struct {
	EnableTrafficLimit      bool
	TrafficLimitMB          int64  `json:"omitempty"`
	TrafficLimitKickMessage string `json:"omitempty"`
}