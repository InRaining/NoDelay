package minecraft

import (
	"fmt"
	"time"
	"strings"

	"github.com/InRaining/NoDelay/common/mcprotocol"
	"github.com/InRaining/NoDelay/config"
	"github.com/InRaining/NoDelay/service/traffic"
)

func generateKickMessage(s *config.ConfigProxyService, name string) mcprotocol.Message {
	return mcprotocol.Message{
		Color: mcprotocol.White,
		Extra: []mcprotocol.Message{
			{Bold: true, Color: mcprotocol.Yellow, Text: fmt.Sprintf("%s", config.Config.Configuration.Header)},
			{Text: " ‖ "},
			{Bold: true, Color: mcprotocol.Red, Text: "已拒绝服务\n"},

			{Text: "您无法加入当前服务器！\n"},
			{Text: "理由: "},
			{Color: mcprotocol.LightPurple, Text: "你的连接可能未经处理，或者你没有权限加入此服务器。\n"},
			{Text: "请联系管理员寻求帮助！\n\n"},

			{
				Color: mcprotocol.Gray,
				Text: fmt.Sprintf("时间戳: %d | 玩家名称: %s | 服务节点: %s\n",
					time.Now().UnixMilli(), name, s.Name),
			},
			{Text: fmt.Sprintf("%s", config.Config.Configuration.ContactName)},
			{Text: ":"},
			{
				Color: mcprotocol.Blue, UnderLined: true,
				Text: fmt.Sprintf("%s", config.Config.Configuration.ContactLink),
				// ClickEvent: chat.OpenURL("http://qm.qq.com/cgi-bin/qm/qr?_wv=1027&k=eV_W6FV6hkjbeA35MNJ2lulA7M67JMig&authKey=E5hHr6NTSJ9u9z7eurOavBW9U6tE94P1EazZSGMGV71LCjsfvgMt0kXRaXyaDF4d&noverify=0&group_code=666259678"),
			},
		},
	}
}

func generatePlayerNumberLimitExceededMessage(s *config.ConfigProxyService, name string) mcprotocol.Message {
	return mcprotocol.Message{
		Color: mcprotocol.White,
		Extra: []mcprotocol.Message{
			{Bold: true, Color: mcprotocol.Yellow, Text: fmt.Sprintf("%s", config.Config.Configuration.Header)},
			{Text: " ‖ "},
			{Bold: true, Color: mcprotocol.Red, Text: "已拒绝服务\n"},

			{Text: "你无法加入当前服务器！\n"},
			{Text: "理由: "},
			{Color: mcprotocol.LightPurple, Text: "服务器当前人数已满载！\n"},
			{Text: "请联系管理员寻求帮助！\n\n"},

			{
				Color: mcprotocol.Gray,
				Text: fmt.Sprintf("时间戳: %d | 玩家名称: %s | 服务节点: %s\n",
					time.Now().UnixMilli(), name, s.Name),
			},
			{Text: fmt.Sprintf("%s", config.Config.Configuration.ContactName)},
			{Text: ":"},
			{
				Color: mcprotocol.Blue, UnderLined: true,
				Text: fmt.Sprintf("%s", config.Config.Configuration.ContactLink),
				// ClickEvent: chat.OpenURL("http://qm.qq.com/cgi-bin/qm/qr?_wv=1027&k=eV_W6FV6hkjbeA35MNJ2lulA7M67JMig&authKey=E5hHr6NTSJ9u9z7eurOavBW9U6tE94P1EazZSGMGV71LCjsfvgMt0kXRaXyaDF4d&noverify=0&group_code=666259678"),
			},
		},
	}
}

func generateNewMessage(s *config.ConfigProxyService, name string) mcprotocol.Message {
	return mcprotocol.Message{
		Color: mcprotocol.White,
		Extra: []mcprotocol.Message{
			{Bold: true, Color: mcprotocol.Green, Text: "=========首次进入提示=========\n"},
			{Color: mcprotocol.Red, Text: "检测到您当前第一次进入本IP!\n"},
			{Color: mcprotocol.LightPurple, Text: "本IP暂不支持防安全警报。\n"},
			{Color: mcprotocol.Blue, Text: "请使用21+或已经历安全警报的账号进入本IP!\n"},
			{Color: mcprotocol.Gold, Text: "一旦被安全警报我们概不负责!\n"},
			{Color: mcprotocol.Green, Text: "如果已经使用21+或已经历安全警报的账号，请尝试重新进入。\n"},
			{Color: mcprotocol.White, Text: "还有其他问题，请开票获取支持!"},
		},
	}
}

func generateJokeMessage(s *config.ConfigProxyService, name string) mcprotocol.Message {
	banID := generateRandomStringWithCharset(8, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	return mcprotocol.Message{
		Color: mcprotocol.White,
		Extra: []mcprotocol.Message{
			{Bold: true, Color: mcprotocol.Red, Text: "You are permanently banned from this server!\n\n"},

			{Color: mcprotocol.Gray, Text: "Reason: "},
			{Text: "Suspicious activity has been detected on your account.\n"},
			{Color: mcprotocol.Gray, Text: "Find out more: "},
			{Color: mcprotocol.Aqua, UnderLined: true, Text: "https://hypixel.net/security\n\n"},
			{Color: mcprotocol.Gray, Text: "Ban ID: "},
			{
				Text: fmt.Sprintf("#%s\n", banID),
			},
			{Color: mcprotocol.Gray, Text: "Sharing your Ban ID may affect the processing of your appeal!"},
		},
	}
}

func generateRandomStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[int(time.Now().UnixNano()+int64(i))%len(charset)]
	}
	return string(b)
}

func generateDownMessage(s *config.ConfigProxyService, name string) mcprotocol.Message {
	return mcprotocol.Message{
		Color: mcprotocol.White,
		Extra: []mcprotocol.Message{
			{Bold: true, Color: mcprotocol.Yellow, Text: fmt.Sprintf("%s", config.Config.Configuration.Header)},
			{Text: " ‖ "},
			{Bold: true, Color: mcprotocol.Gold, Text: "已拒绝服务\n"},

			{Text: "您无法加入当前服务器！\n"},
			{Text: "理由: "},
			{Color: mcprotocol.LightPurple, Text: "当前正在进行停机维护！\n"},
			{Text: "请关注相关信息渠道了解恢复时间！\n\n"},

			{
				Color: mcprotocol.Gray,
				Text: fmt.Sprintf("时间戳: %d | 玩家名称: %s | 服务节点: %s\n",
					time.Now().UnixMilli(), name, s.Name),
			},
			{Text: fmt.Sprintf("%s", config.Config.Configuration.ContactName)},
			{Text: ":"},
			{
				Color: mcprotocol.Blue, UnderLined: true,
				Text: fmt.Sprintf("%s", config.Config.Configuration.ContactLink),
				// ClickEvent: chat.OpenURL("http://qm.qq.com/cgi-bin/qm/qr?_wv=1027&k=eV_W6FV6hkjbeA35MNJ2lulA7M67JMig&authKey=E5hHr6NTSJ9u9z7eurOavBW9U6tE94P1EazZSGMGV71LCjsfvgMt0kXRaXyaDF4d&noverify=0&group_code=666259678"),
			},
		},
	}
}

func generateTrafficLimitExceededMessage(s *config.ConfigProxyService, name string) mcprotocol.Message {
    used, limit, percentage := traffic.GetUserTrafficInfoByPlayer(name)

    if config.Config.TrafficLimiter.TrafficLimitKickMessage != "" {
        message := config.Config.TrafficLimiter.TrafficLimitKickMessage
        message = strings.ReplaceAll(message, "{player}", name)
        message = strings.ReplaceAll(message, "{used}", fmt.Sprintf("%.2f", used))
        message = strings.ReplaceAll(message, "{limit}", fmt.Sprintf("%.0f", limit))
        message = strings.ReplaceAll(message, "{percentage}", fmt.Sprintf("%.1f", percentage))
        return mcprotocol.Message{Text: message}
    }

    return mcprotocol.Message{
        Color: mcprotocol.White,
        Extra: []mcprotocol.Message{
			{Bold: true, Color: mcprotocol.Yellow, Text: fmt.Sprintf("%s", config.Config.Configuration.Header)},
			{Text: " ‖ "},
			{Bold: true, Color: mcprotocol.Gold, Text: "已拒绝服务\n"},

			{Text: "您无法加入当前服务器！\n"},
			{Text: "理由: "},
			{Color: mcprotocol.LightPurple, Text: "流量已耗尽！\n"},

            {Color: mcprotocol.Gray, Text: "已使用: "},
            {Color: mcprotocol.Yellow, Text: fmt.Sprintf("%.2f MB ", used)},
            {Color: mcprotocol.Gray, Text: "/ "},
            {Color: mcprotocol.Green, Text: fmt.Sprintf("%.0f MB ", limit)},
            {Color: mcprotocol.White, Text: fmt.Sprintf("(%.1f%%)\n", percentage)},
			{Text: "请联系管理员寻求帮助！\n\n"},
			{
				Color: mcprotocol.Gray,
				Text: fmt.Sprintf("时间戳: %d | 玩家名称: %s | 服务节点: %s\n",
					time.Now().UnixMilli(), name, s.Name),
			},
			{Text: fmt.Sprintf("%s", config.Config.Configuration.ContactName)},
			{Text: ":"},
			{
				Color: mcprotocol.Blue, UnderLined: true,
				Text: fmt.Sprintf("%s", config.Config.Configuration.ContactLink),
				// ClickEvent: chat.OpenURL("http://qm.qq.com/cgi-bin/qm/qr?_wv=1027&k=eV_W6FV6hkjbeA35MNJ2lulA7M67JMig&authKey=E5hHr6NTSJ9u9z7eurOavBW9U6tE94P1EazZSGMGV71LCjsfvgMt0kXRaXyaDF4d&noverify=0&group_code=666259678"),
			},
        },
    }
}
