package op

func Name(op Operation) (string, bool) {
	switch v := op.(type) {
	case Open:
		return v.Name, true
	case Glob:
		return v.Pattern, true
	case Lstat:
		return v.Name, true
	case ReadDir:
		return v.Name, true
	case ReadFile:
		return v.Name, true
	case WriteFile:
		return v.Name, true
	case Remove:
		return v.Name, true
	case RemoveAll:
		return v.Name, true
	default:
		return "", false
	}
}
