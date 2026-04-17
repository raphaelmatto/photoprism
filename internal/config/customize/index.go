package customize

// IndexSettings represents indexing settings.
type IndexSettings struct {
	Path          string `json:"path" yaml:"Path"`
	Convert       bool   `json:"convert" yaml:"Convert"`
	Rescan        bool   `json:"rescan" yaml:"Rescan"`
	SkipArchived  bool   `json:"skipArchived" yaml:"SkipArchived"`
	AddAIKeywords *bool  `json:"addAIKeywords" yaml:"AddAIKeywords"`
}

// AIKeywordsEnabled reports whether AI-generated labels should be converted into searchable keywords.
func (s IndexSettings) AIKeywordsEnabled() bool {
	return s.AddAIKeywords == nil || *s.AddAIKeywords
}

func newBool(value bool) *bool {
	v := value
	return &v
}
