package factory

type Fs struct {
	name string
}

func NewFs() *Fs {
	return &Fs{}
}

func (f *Fs) Named(name string) *Fs {
	f.name = name
	return f
}
