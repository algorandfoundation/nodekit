package overlay

type Modal interface {
	Title() string
	BorderColor() string
	Controls() string
	Body() string
}
