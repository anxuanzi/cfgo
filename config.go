package cfgo

type Config interface {
	Get(string) string
	GetOrDefault(string, string) string
	GetArray(string) []string
}
