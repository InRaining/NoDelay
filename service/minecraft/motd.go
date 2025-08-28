package minecraft

import (
	"encoding/json"

	"github.com/CubeWhyMC/NoDelay-Proxy-Server/config"
	"github.com/CubeWhyMC/NoDelay-Proxy-Server/service/transfer"
	"github.com/CubeWhyMC/NoDelay-Proxy-Server/version"
)

type motdObject struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Players struct {
		Max    int `json:"max"`
		Online int `json:"online"`
		Sample any `json:"sample,omitempty"`
	} `json:"players"`
	Description struct {
		Text string `json:"text"`
	} `json:"description"`
	Favicon string `json:"favicon"`
}

func generateMOTD(protocolVersion int, s *config.ConfigProxyService, options *transfer.Options) []byte {
	online := s.Minecraft.OnlineCount.Online
	if online < 0 {
		online = options.OnlineCount.Load()
	}

	motd, _ := json.Marshal(motdObject{
		Version: struct {
			Name     string `json:"name"`
			Protocol int    `json:"protocol"`
		}{
			Name:     "NoDelay " + version.Version,
			Protocol: protocolVersion,
		},
		Players: struct {
			Max    int `json:"max"`
			Online int `json:"online"`
			Sample any `json:"sample,omitempty"`
		}{
			Max:    s.Minecraft.OnlineCount.Max,
			Online: int(online),
			Sample: s.Minecraft.OnlineCount.Sample,
		},
		Description: struct {
			Text string `json:"text"`
		}{
			Text: s.Minecraft.MotdDescription,
		},
		Favicon: s.Minecraft.MotdFavicon,
	})

	return motd
}
