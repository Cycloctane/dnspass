package records

var defaultStore = NewStore()

func SetAll(records []Record) {
	defaultStore.SetAll(records)
}

func Add(r Record) bool {
	return defaultStore.Add(r)
}

func Delete(name string, t DNSType, value string) bool {
	return defaultStore.Delete(name, t, value)
}

func Lookup(name string, t DNSType) []Record {
	return defaultStore.Get(name, t)
}

func Lookup1(name string, t DNSType) (Record, bool) {
	return defaultStore.Get1(name, t)
}

func List() []Record {
	return defaultStore.List()
}
