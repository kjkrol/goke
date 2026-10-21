package query

// walk is a Walker over m, standing nowhere
func walk(m *Matcher) *Walker {
	var w Walker
	w.Init(m)
	return &w
}
