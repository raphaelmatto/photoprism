package customize

// FilterSettings controls which filter dropdowns appear in the photo search
// toolbar. Each field is a pointer so that omission in the YAML file can be
// distinguished from an explicit false; Propagate() fills in nil pointers
// with the default (visible).
type FilterSettings struct {
	Country  *bool `json:"country" yaml:"Country,omitempty"`
	Camera   *bool `json:"camera" yaml:"Camera,omitempty"`
	View     *bool `json:"view" yaml:"View,omitempty"`
	Order    *bool `json:"order" yaml:"Order,omitempty"`
	Year     *bool `json:"year" yaml:"Year,omitempty"`
	Month    *bool `json:"month" yaml:"Month,omitempty"`
	Color    *bool `json:"color" yaml:"Color,omitempty"`
	Category *bool `json:"category" yaml:"Category,omitempty"`
}

// NewFilterSettings returns FilterSettings with every filter visible by
// default, preserving the historical behavior for fresh installs and existing
// users that had no filters key in their settings.yml.
func NewFilterSettings() FilterSettings {
	return FilterSettings{
		Country:  newBool(true),
		Camera:   newBool(true),
		View:     newBool(true),
		Order:    newBool(true),
		Year:     newBool(true),
		Month:    newBool(true),
		Color:    newBool(true),
		Category: newBool(true),
	}
}
