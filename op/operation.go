package op

type Open struct {
	Name string
}

func (o Open) Path() string {
	return o.Name
}
