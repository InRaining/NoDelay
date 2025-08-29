package service

import (
	"log"
	"net"
	"strconv"

	"github.com/InRaining/NoDelay/common"
	"github.com/InRaining/NoDelay/config"
	"github.com/InRaining/NoDelay/service/minecraft"
	"github.com/InRaining/NoDelay/service/tls"
	"github.com/InRaining/NoDelay/service/transfer"

	"github.com/fatih/color"
)

var (
	GreenPlus = color.HiGreenString("[+]")
	RedMinus  = color.HiRedString("[-]")
)

func newConnReceiver(s *config.ConfigProxyService,
	conn *net.TCPConn,
	options *transfer.Options,
) {
	ctx := new(transfer.ConnContext).Init()
	log.Println("Service", s.Name, ":", ctx.ColoredID, GreenPlus, conn.RemoteAddr().String())
	defer log.Println("Service", s.Name, ":", ctx.ColoredID, RedMinus, conn.RemoteAddr().String(), ctx)
	var remote net.Conn

	if options.IsTLSHandleNeeded {
		remote, ctx.Err = tls.NewConnHandler(s, conn, options.Out)
		if ctx.Err != nil {
			conn.Close()
			return
		}
	} else if options.IsMinecraftHandleNeeded {
		remote, ctx.Err = minecraft.NewConnHandler(s, ctx, conn, options)
		if ctx.Err != nil {
			conn.Close()
			return
		}
	}

	if remote == nil {
		var err error
		remote, err = options.Out.Dial("tcp", net.JoinHostPort(s.TargetAddress, strconv.FormatInt(int64(s.TargetPort), 10)))
		if err != nil {
			ctx.Err = common.Cause("failed to dial to target server: ", err)
			conn.Close()
			return
		}
	}
	options.OnlineCount.Add(1)
	defer options.OnlineCount.Add(-1)
	transfer.SimpleTransfer(conn, remote, options.FlowType)
}
