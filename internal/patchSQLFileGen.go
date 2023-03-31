package internal

import (
	"context"
	"errors"
	"os"

	"github.com/elgris/sqrl"
	"github.com/google/wire"
	interpolater "github.com/huandu/go-sqlbuilder"
	"github.com/tinhtran24/xo-patcher/utils"
)

type PatchSQLFileGen struct {
	file *os.File
}

type IPatchSQLFileGen interface {
	Write(sqlizer sqrl.Sqlizer) error
	Close()
}

var NewPatchSQLFileGen = wire.NewSet(
	InitPatchSQLFileGen,
	wire.Bind(new(IPatchSQLFileGen), new(PatchSQLFileGen)),
)

func InitPatchSQLFileGen(ctx context.Context) (*PatchSQLFileGen, error) {
	patchName, err := getPatchContext(ctx)
	if err != nil {
		return nil, err
	}
	file, err := os.Create("patches_gen/" + patchName + ".sql")
	if err != nil {
		return nil, err
	}
	return &PatchSQLFileGen{file: file}, nil

}

func (fileGen *PatchSQLFileGen) Write(sqlizer sqrl.Sqlizer) error {
	query, args, err := sqlizer.ToSql()
	if err != nil {
		return err
	}
	query, err = interpolater.MySQL.Interpolate(query, args)
	if err != nil {
		return err
	}
	_, err = fileGen.file.WriteString(query + ";\n")
	if err != nil {
		return err
	}
	return nil

}

func (fileGen *PatchSQLFileGen) Close() {
	fileGen.file.Close()
}

func getPatchContext(ctx context.Context) (string, error) {
	if value, ok := ctx.Value(utils.PatchName).(string); ok {
		return value, nil
	}
	return "", errors.New("patchName context invalid")
}
