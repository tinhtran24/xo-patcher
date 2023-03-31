package internal

type FilterType string

const (
	Eq      FilterType = "eq"
	Neq     FilterType = "Neq"
	Gt      FilterType = "gt"
	Gte     FilterType = "gte"
	Lt      FilterType = "lt"
	Lte     FilterType = "lte"
	Like    FilterType = "like"
	Between FilterType = "between"
)

// type FilterType int

// const (
// 	Eq      FilterType = iota
// 	Neq
// 	Gt
// 	Gte
// 	Lt
// 	Lte
// 	Like
// 	Between
// )

// func (ft *FilterType) String() string {
// 	switch *ft {
// 	case Eq:
// 		return "eq"
// 	}
// 	...
// 	return ""
// }

// func (ft *FilterType) UnMarshal(value string) err {
// 	switch value {
// 	case "eq":
// 		*ft = Eq
// 		return nil
// 	}
// 	...
// 	return errors.New("invalid value")
// }
