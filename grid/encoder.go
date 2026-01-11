package grid

type Encoder interface {
	EncodeSheet(View) error
}
