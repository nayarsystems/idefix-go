package main

import (
	"context"

	"github.com/alecthomas/kong"
	idf "gitlab.com/garagemakers/idefix-go"
)

type Context struct {
	Client  *idf.Client
	Context context.Context
	Cancel  context.CancelFunc
}

var cli struct {
	ConfigName string `help:"Idefix client config filename" short:"c" default:"default"`

	Log  LogCmd  `cmd:"" help:"Stream device log"`
	Info InfoCmd `cmd:"" help:"Get device info"`
}

func main() {
	var err error
	ctx := kong.Parse(&cli)

	kctx := &Context{}
	kctx.Context, kctx.Cancel = context.WithCancel(context.Background())
	kctx.Client, err = idf.NewClientFromFile(kctx.Context, cli.ConfigName)
	if err != nil {
		ctx.FatalIfErrorf(err)
	}

	ctx.FatalIfErrorf(ctx.Run(kctx))
}
