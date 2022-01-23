package hash

type Hash interface {
	Generate(password string) (string, error)
	Check(hash, password string) error
}
