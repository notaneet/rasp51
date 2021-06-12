package converter

func Converter(converter string) IConverter {
	switch converter {
	case "json":
		return JSONConverter{}
	case "pjson":
		return JSONConverter{Pretty: true}
	case "pgsql":
		return PGSQLConverter{}
	default:
		return DummyConverter{}
	}
}
