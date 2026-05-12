package photoprism

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/entity"
)

// metadataSrcColumns lists the photo and details *_src columns that PhotoPrism
// uses to track where a value came from. SetTitle, SetKeywords, SetCaption and
// their siblings compare the priority of the existing source to the new one
// before overwriting, so a row marked manual or batch will refuse subsequent
// file-metadata updates. ResetMetadataSrc clears those markers so a re-index
// can repopulate values from the latest IPTC, EXIF, and XMP data.
var metadataSrcColumns = []struct {
	Table  string
	Column string
}{
	{"photos", "taken_src"},
	{"photos", "type_src"},
	{"photos", "title_src"},
	{"photos", "caption_src"},
	{"photos", "place_src"},
	{"photos", "camera_src"},
	{"details", "keywords_src"},
	{"details", "notes_src"},
	{"details", "subject_src"},
	{"details", "artist_src"},
	{"details", "copyright_src"},
	{"details", "license_src"},
	{"details", "software_src"},
}

// ResetMetadataSrc clears the manual and batch source markers on the photo
// and details tables. After this call the next index pass treats file
// metadata as having higher priority than the cleared rows, so re-exported
// IPTC/EXIF/XMP values overwrite previously-edited fields. Returns the
// number of rows updated across all targeted columns.
func ResetMetadataSrc() (int64, error) {
	db := entity.UnscopedDb()
	if db == nil {
		return 0, fmt.Errorf("reset metadata src: database unavailable")
	}

	var total int64
	for _, c := range metadataSrcColumns {
		stmt := fmt.Sprintf("UPDATE %s SET %s = '' WHERE %s IN ('manual', 'batch')", c.Table, c.Column, c.Column)
		res := db.Exec(stmt)
		if res.Error != nil {
			return total, fmt.Errorf("reset metadata src: %s.%s: %w", c.Table, c.Column, res.Error)
		}
		total += res.RowsAffected
	}

	return total, nil
}
