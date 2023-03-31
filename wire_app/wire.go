//go:build wireinject
// +build wireinject

package wire_app

import (
	"context"
	"github.com/google/wire"

	"github.com/tinhtran24/xo-patcher/internal"
	"github.com/tinhtran24/xo/xo_wire"
)

// To inject all patch to App
// This will allow calls to InitPatch method
// and we can also we can directly call run method of individual patch

var patchSet = wire.NewSet(
	patches.NewOne,
)

type PatchGroup struct {
	One patches.IOne
}

type App struct {
	PatchGroup   PatchGroup
	PatchManager internal.IPatchManager
}

var globalSet = wire.NewSet(
	xo_wire.RepositorySet,
	wire.Struct(new(App), "*"),
	wire.Struct(new(PatchGroup), "*"),
	internal.NewPatchManager,
	patchSet,
	internal.NewDB,
	internal.NewPatchSQLFileGen,
)

func GetApp(ctx context.Context) (*App, func(), error) {
	wire.Build(globalSet)
	return &App{}, nil, nil
}
