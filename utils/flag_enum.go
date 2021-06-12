package utils

type StringEnum []string

func (i *StringEnum) String() string {
	return ""
}

func (i *StringEnum) Set(value string) error {
	*i = append(*i, value)
	return nil
}
