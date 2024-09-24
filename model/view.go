package model

type View struct {
	Name  string     `mapstructure:"name"`
	Facts []ViewFact `mapstructure:"facts"`
}

type ViewFact struct {
	Name     string `mapstructure:"name"`
	Fact     string `mapstructure:"fact"`
	Renderer string `mapstructure:"renderer"`
}

type ViewResult struct {
	View View
	Data any
}
