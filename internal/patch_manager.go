package internal

import (
	"context"
	"errors"

	"github.com/google/wire"
)

type IPatchRunner interface {
	Run(ctx context.Context) error
}

type IPatchManager interface {
	Run(ctx context.Context, name string) error
	RegisterPatch(name string, runner IPatchRunner) error
}

type PatchManageOptions struct {
	DB      IDb
	FileGen IPatchSQLFileGen
}

type PatchManager struct {
	*PatchManageOptions
	patchList map[string]IPatchRunner
}

func InitPatchManager(options *PatchManageOptions) *PatchManager {
	return &PatchManager{
		PatchManageOptions: options,
		patchList:          map[string]IPatchRunner{},
	}
}

var NewPatchManager = wire.NewSet(
	wire.Struct(new(PatchManageOptions), "*"),
	InitPatchManager,
	wire.Bind(new(IPatchManager), new(PatchManager)),
)

func (pm *PatchManager) RegisterPatch(name string, runner IPatchRunner) error {
	if _, ok := pm.patchList[name]; ok {
		return errors.New("patch Already exists")
	}
	pm.patchList[name] = runner
	return nil
}

func (pm *PatchManager) Run(ctx context.Context, name string) error {
	// close file after run completed
	defer pm.FileGen.Close()
	if runner, ok := pm.patchList[name]; ok {
		return WrapInTransaction(ctx, pm.DB, func(ctx context.Context) error {
			return runner.Run(ctx)
		})
	}
	return errors.New("patch by name: " + name + " not found")
}
