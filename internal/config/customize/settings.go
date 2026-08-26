package customize

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// RootPath defines the default root directory used for import and index settings.
const (
	RootPath = "/"
)

// Settings represents user settings for Web UI, indexing, and import.
type Settings struct {
	UI        UISettings       `json:"ui" yaml:"UI"`
	Search    SearchSettings   `json:"search" yaml:"Search"`
	Maps      MapsSettings     `json:"maps" yaml:"Maps"`
	Features  FeatureSettings  `json:"features" yaml:"Features"`
	Import    ImportSettings   `json:"import" yaml:"Import"`
	Index     IndexSettings    `json:"index" yaml:"Index"`
	Stack     StackSettings    `json:"stack" yaml:"Stack"`
	Share     ShareSettings    `json:"share" yaml:"Share"`
	Download  DownloadSettings `json:"download" yaml:"Download"`
	Display   DisplaySettings  `json:"display" yaml:"Display"`
	Albums    AlbumsSettings   `json:"albums" yaml:"Albums"`
	Templates TemplateSettings `json:"templates" yaml:"Templates"`
}

// NewDefaultSettings creates a new default Settings instance.
func NewDefaultSettings() *Settings {
	return NewSettings(DefaultTheme, DefaultLanguage, DefaultTimeZone)
}

// NewSettings creates a new Settings instance.
func NewSettings(theme, language, timeZone string) *Settings {
	if theme == "" {
		theme = DefaultTheme
	}

	if language == "" {
		language = DefaultLanguage
	}

	if timeZone == "" {
		timeZone = DefaultTimeZone
	}

	return &Settings{
		UI: UISettings{
			Scrollbar: true,
			Zoom:      false,
			Theme:     theme,
			Language:  language,
			TimeZone:  timeZone,
			StartPage: DefaultStartPage,
		},
		Search: SearchSettings{
			BatchSize:    -1,
			ListView:     true,
			ShowTitles:   true,
			ShowCaptions: true,
		},
		Maps: MapsSettings{
			Animate: 0,
			Style:   DefaultMapsStyle,
		},
		Features: NewFeatures(),
		Import: ImportSettings{
			Path: RootPath,
			Move: false,
			Dest: "",
		},
		Index: IndexSettings{
			Path:          RootPath,
			Rescan:        false,
			Convert:       true,
			AddAIKeywords: newBool(true),
		},
		Stack: StackSettings{
			UUID: true,
			Meta: true,
			Name: false,
		},
		Share: ShareSettings{
			Title: "",
		},
		Albums:   NewAlbumSettings(),
		Download: NewDownloadSettings(),
		Display: DisplaySettings{
			Originals:        false,
			ImagePacking:     false,
			LightboxBorder:   1,
			RetinaLightbox:   false,
			RetinaThumbnails: false,
			Metadata:         NewMetadataLayoutSettings(),
			Filters:          NewFilterSettings(),
			Navigation:       NewNavigationSettings(),
		},
		Templates: TemplateSettings{
			Default: "index.gohtml",
		},
	}
}

// Propagate updates settings in other packages as needed.
func (s *Settings) Propagate() {
	if s.UI.Language == "" {
		s.UI.Language = DefaultLanguage
	}

	if s.UI.TimeZone == "" {
		s.UI.TimeZone = DefaultTimeZone
	}

	if s.UI.StartPage == "" {
		s.UI.StartPage = DefaultStartPage
	}

	if s.Index.AddAIKeywords == nil {
		s.Index.AddAIKeywords = newBool(true)
	}

	if s.Maps.Style == "" {
		s.Maps.Style = DefaultMapsStyle
	}

	// Migrate the legacy folder default so existing settings files use capture-date order.
	if s.Albums.Order.Folder == "" || s.Albums.Order.Folder == sortby.Added {
		s.Albums.Order.Folder = sortby.Oldest
	}

	defaultMetadata := NewMetadataLayoutSettings()

	if s.Display.Metadata.Cards == nil {
		s.Display.Metadata.Cards = defaultMetadata.Cards
	}

	if s.Display.Metadata.List == nil {
		s.Display.Metadata.List = defaultMetadata.List
	}

	if s.Display.Metadata.Lightbox == nil {
		s.Display.Metadata.Lightbox = defaultMetadata.Lightbox
	}

	if s.Display.LightboxBorder < 0 {
		s.Display.LightboxBorder = 0
	}

	// Fill in nil filter visibility pointers with the default (visible) so
	// existing settings.yml files without a Filters section keep working.
	defaultFilters := NewFilterSettings()
	if s.Display.Filters.Country == nil {
		s.Display.Filters.Country = defaultFilters.Country
	}
	if s.Display.Filters.Camera == nil {
		s.Display.Filters.Camera = defaultFilters.Camera
	}
	if s.Display.Filters.Lens == nil {
		s.Display.Filters.Lens = defaultFilters.Lens
	}
	if s.Display.Filters.View == nil {
		s.Display.Filters.View = defaultFilters.View
	}
	if s.Display.Filters.Order == nil {
		s.Display.Filters.Order = defaultFilters.Order
	}
	if s.Display.Filters.Year == nil {
		s.Display.Filters.Year = defaultFilters.Year
	}
	if s.Display.Filters.Month == nil {
		s.Display.Filters.Month = defaultFilters.Month
	}
	if s.Display.Filters.Color == nil {
		s.Display.Filters.Color = defaultFilters.Color
	}
	if s.Display.Filters.Category == nil {
		s.Display.Filters.Category = defaultFilters.Category
	}

	// Fill in nil navigation visibility pointers with the default (visible)
	// so existing settings.yml files without a Navigation block keep their
	// sidebar entries.
	defaultNav := NewNavigationSettings()
	if s.Display.Navigation.Albums == nil {
		s.Display.Navigation.Albums = defaultNav.Albums
	}
	if s.Display.Navigation.Media == nil {
		s.Display.Navigation.Media = defaultNav.Media
	}
	if s.Display.Navigation.Unsorted == nil {
		s.Display.Navigation.Unsorted = defaultNav.Unsorted
	}
	if s.Display.Navigation.SearchFilters == nil {
		s.Display.Navigation.SearchFilters = defaultNav.SearchFilters
	}

	i18n.SetLocale(s.UI.Language)
}

// StackSequences checks if files should be stacked based on their file name prefix (sequential names).
func (s Settings) StackSequences() bool {
	return s.Stack.Name
}

// StackUUID checks if files should be stacked based on unique image or instance id.
func (s Settings) StackUUID() bool {
	return s.Stack.UUID
}

// StackMeta checks if files should be stacked based on their place and time metadata.
func (s Settings) StackMeta() bool {
	return s.Stack.Meta
}

// Load user settings from file.
func (s *Settings) Load(fileName string) error {
	if fileName == "" {
		return fmt.Errorf("missing settings filename")
	} else if !fs.FileExists(fileName) {
		return fmt.Errorf("settings file not found: %s", clean.Log(fileName))
	}

	name := filepath.Clean(fileName)

	yamlConfig, err := os.ReadFile(name) // #nosec G304 -- file path is provided by the caller and validated above

	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(yamlConfig, s); err != nil {
		return err
	}

	s.Propagate()

	return nil
}

// Save user settings to a file.
func (s *Settings) Save(fileName string) error {
	if fileName == "" {
		return fmt.Errorf("missing settings filename")
	}

	data, err := yaml.Marshal(s)

	if err != nil {
		return err
	}

	s.Propagate()

	return os.WriteFile(fileName, data, fs.ModeConfigFile)
}
