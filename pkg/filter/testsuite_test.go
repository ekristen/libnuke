package filter_test

type TestResource struct{}

func (t *TestResource) GetProperty(key string) (string, error) {
	return "testing", nil
}
