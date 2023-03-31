package main

import (
	"context"
	"flag"

	"github.com/tinhtran24/xo-patcher/utils"
	"github.com/tinhtran24/xo-patcher/wire_app"
)

var patchNameFlag = flag.String("patch-name", "", "name of patch to run")
var commitFlag = flag.Bool("commit", false, "commit update")
var prodFlag = flag.Bool("prod", false, "skip exec")
var connectionStringFlag = flag.String("connection", "", "database connection string")

func main() {

	flag.Parse()

	connectionString := *connectionStringFlag
	patchName := *patchNameFlag
	if connectionString == "" {
		panic("connection flag must be provided")
	}
	if patchName == "" {
		panic("patch-name flag must be provided")
	}
	commit := *commitFlag
	prod := *prodFlag

	ctx := context.WithValue(context.Background(), utils.Connection, connectionString)
	ctx = context.WithValue(ctx, utils.Commit, commit)
	ctx = context.WithValue(ctx, utils.Prod, prod)
	ctx = context.WithValue(ctx, utils.PatchName, patchName)

	app, _, err := wire_app.GetApp(ctx)

	if err != nil {
		panic(err)
	}

	err = app.PatchManager.Run(ctx, patchName)
	if err != nil {
		panic(err)
	}

}
