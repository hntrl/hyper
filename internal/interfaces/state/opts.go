package state

import "github.com/hntrl/hyper/internal/symbols"

type QueryOpts struct {
	Skip  *symbols.IntegerLiteral
	Limit *symbols.IntegerLiteral
}

func (qo QueryOpts) ClassName() string {
	return "QueryOpts"
}
func (qo QueryOpts) Fields() map[string]symbols.Class {
	return map[string]symbols.Class{
		"skip":  symbols.NewOptionalClass(symbols.Integer{}),
		"limit": symbols.NewOptionalClass(symbols.Integer{}),
	}
}
func (qo QueryOpts) Constructors() symbols.ConstructorMap {
	csMap := symbols.NewConstructorMap()
	csMap.AddGenericConstructor(qo, func(fields map[string]symbols.ValueObject) (symbols.ValueObject, error) {
		qo := QueryOpts{}
		if skip, ok := fields["skip"]; ok {
			// LMAO this is so bad
			obj := skip.(*symbols.NilableObject).Object
			if obj != nil {
				lit := obj.(symbols.IntegerLiteral)
				qo.Skip = &lit
			}
		}
		if limit, ok := fields["limit"]; ok {
			obj := limit.(*symbols.NilableObject).Object
			if obj != nil {
				lit := obj.(symbols.IntegerLiteral)
				qo.Limit = &lit
			}
		}
		return &qo, nil
	})
	return csMap
}

func (qo QueryOpts) Class() symbols.Class {
	return qo
}
func (qo QueryOpts) Value() interface{} {
	return nil
}
func (qo *QueryOpts) Set(key string, obj symbols.ValueObject) error {
	return nil
}
func (qo QueryOpts) Get(key string) (symbols.Object, error) {
	return nil, nil
}
