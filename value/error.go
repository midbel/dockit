package value

type ErrorCode string

var (
	ErrNull  = createError("#NULL!")
	ErrDiv0  = createError("#DIV/0!")
	ErrValue = createError("#VALUE!")
	ErrRef   = createError("#REF!")
	ErrName  = createError("#NAME?")
	ErrNum   = createError("#NUM!")
	ErrNA    = createError("#N/A")
)

type Error struct {
	code string
}

func createError(code string) Error {
	return Error{
		code: code,
	}
}

func (Error) Type() string {
	return "error"
}

func (Error) Kind() ValueKind {
	return KindError
}

func (e Error) Error() string {
	return e.code
}

func (e Error) String() string {
	return e.code
}

func (e Error) Scalar() any {
	return e.code
}
