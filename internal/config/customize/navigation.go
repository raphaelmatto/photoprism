package customize

// NavigationSettings controls which entries appear in the sidebar
// navigation. Each field is a pointer so omission in the YAML file can be
// distinguished from an explicit false; Propagate() fills in nil pointers
// with the default (visible). Existing FeatureSettings flags like
// features.albums and features.videos still apply on top of these — both
// must be true for an entry to render.
type NavigationSettings struct {
	Albums        *bool `json:"albums" yaml:"Albums,omitempty"`
	Media         *bool `json:"media" yaml:"Media,omitempty"`
	Unsorted      *bool `json:"unsorted" yaml:"Unsorted,omitempty"`
	SearchFilters *bool `json:"searchFilters" yaml:"SearchFilters,omitempty"`
}

// NewNavigationSettings returns NavigationSettings with every sidebar entry
// visible by default, preserving the historical behavior for fresh installs
// and existing users whose settings.yml has no Navigation block.
func NewNavigationSettings() NavigationSettings {
	return NavigationSettings{
		Albums:        newBool(true),
		Media:         newBool(true),
		Unsorted:      newBool(true),
		SearchFilters: newBool(true),
	}
}
