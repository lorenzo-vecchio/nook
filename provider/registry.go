package provider

var registry = map[string]Provider{}

func Register(p Provider) {
	registry[p.Name()] = p
}

func Get(name string) (Provider, bool) {
	p, ok := registry[name]
	return p, ok
}

func List() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
