package main

import (
	_ "cex-exchange/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"

	"cex-exchange/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
