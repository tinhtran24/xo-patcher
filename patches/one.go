package patches

import (
	"context"

	"github.com/google/wire"
	"github.com/tinhtran24/xo-patcher/internal"
)

type IOne interface {
}

type OneOptions struct {
	PatchManager internal.IPatchManager
}

type One struct {
	*OneOptions
	// and my extra for variables
}

var NewOne = wire.NewSet(
	wire.Struct(new(OneOptions), "*"),
	InitOne,
	wire.Bind(new(IOne), new(One)),
)

func InitOne(options *OneOptions) (*One, error) {
	one := &One{OneOptions: options}
	err := options.PatchManager.RegisterPatch("one", one)
	return one, err
}

func (one *One) Run(ctx context.Context) error {
	return nil
}

/*
class OneImp implements IOne {
	variables ...One
	constructor InitOne(...OneOptions) {
		// custom logic
	}
	func Run() { }
}
*/
