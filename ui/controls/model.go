package controls

// Model represents the data structure used for defining visibility, dimensions, and content.
type Model struct {
	Width     int
	Height    int
	IsVisible bool
	Content   string
}

// New creates a instance of a Model
func New(body string) Model {
	return Model{
		IsVisible: true,
		Width:     80,
		Height:    24,
		Content:   body,
	}
}
