package api

type ResponseInterface interface {
	StatusCode() int
	Status() string
}
