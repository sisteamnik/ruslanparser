package ruslanparser

type Option struct {
	t     Opt
	value string
}

func NewOpt(value string, t Opt) *Option {
	return &Option{t, value}
}

type Opt int

const (
	OptTypeSessionFile Opt = iota
)

func (o *Option) Type() Opt {
	return o.t
}

func (o *Option) Value() string {
	return o.value
}
