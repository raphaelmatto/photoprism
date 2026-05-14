package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/entity"
)

// DumpEditsCommand lists photos whose metadata has been edited in the web
// UI but not yet round-tripped back to the source files (e.g. into Lightroom
// via XMP). It is the "extract pending edits" step in the workflow:
//
//  1. Users edit captions/keywords/etc. in PhotoPrism.
//  2. Run `photoprism dump-edits` (this command) to produce a manifest.
//  3. Apply the edits in Lightroom by hand, or feed the JSON manifest to
//     an LR-side script. Save Metadata to File so XMP sidecars are written.
//  4. Re-index the affected paths with `--overwrite-meta`. This pulls the
//     edits in from XMP and clears the `*_src='manual'|'batch'` markers,
//     so the next dump-edits run only surfaces edits made after this one.
var DumpEditsCommand = &cli.Command{
	Name:   "dump-edits",
	Usage:  "Lists photos whose title, caption, taken-at, keywords, location, or details have been manually edited in the web UI",
	Flags:  dumpEditsFlags,
	Action: dumpEditsAction,
}

var dumpEditsFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:  "json",
		Usage: "emit a structured JSON array instead of a markdown manifest",
	},
}

// editValue captures one manually-edited field for a photo. Src is the
// raw `*_src` column value ('manual' or 'batch'); Value is the post-edit
// payload (string, []string, or a small map for compound values).
type editValue struct {
	Src   string      `json:"src"`
	Value interface{} `json:"value"`
}

// editRow is one photo's worth of manual edits, indexed by field name.
type editRow struct {
	UID       string               `json:"uid"`
	Path      string               `json:"path"`
	Original  string               `json:"original_name,omitempty"`
	UpdatedAt time.Time            `json:"updated_at"`
	Fields    map[string]editValue `json:"fields"`
}

// dumpEditsRaw mirrors the SQL projection one-to-one; the post-processing
// step below turns it into the field map.
type dumpEditsRaw struct {
	PhotoUID         string    `gorm:"column:photo_uid"`
	PhotoPath        string    `gorm:"column:photo_path"`
	PhotoName        string    `gorm:"column:photo_name"`
	OriginalName     string    `gorm:"column:original_name"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
	TakenAtLocal     time.Time `gorm:"column:taken_at_local"`
	TimeZone         string    `gorm:"column:time_zone"`
	PhotoTitle       string    `gorm:"column:photo_title"`
	TitleSrc         string    `gorm:"column:title_src"`
	PhotoCaption     string    `gorm:"column:photo_caption"`
	CaptionSrc       string    `gorm:"column:caption_src"`
	TakenSrc         string    `gorm:"column:taken_src"`
	PhotoLat         float64   `gorm:"column:photo_lat"`
	PhotoLng         float64   `gorm:"column:photo_lng"`
	PlaceSrc         string    `gorm:"column:place_src"`
	DetailsKeywords  string    `gorm:"column:details_keywords"`
	KeywordsSrc      string    `gorm:"column:keywords_src"`
	DetailsSubject   string    `gorm:"column:details_subject"`
	SubjectSrc       string    `gorm:"column:subject_src"`
	DetailsNotes     string    `gorm:"column:details_notes"`
	NotesSrc         string    `gorm:"column:notes_src"`
	DetailsArtist    string    `gorm:"column:details_artist"`
	ArtistSrc        string    `gorm:"column:artist_src"`
	DetailsCopyright string    `gorm:"column:details_copyright"`
	CopyrightSrc     string    `gorm:"column:copyright_src"`
}

func dumpEditsAction(ctx *cli.Context) error {
	conf, err := InitConfig(ctx)
	if err != nil {
		return err
	}

	conf.InitDb()
	defer conf.Shutdown()

	var rows []dumpEditsRaw
	if err = entity.UnscopedDb().Raw(`
		SELECT
			p.photo_uid                          AS photo_uid,
			p.photo_path                         AS photo_path,
			p.photo_name                         AS photo_name,
			p.original_name                      AS original_name,
			p.updated_at                         AS updated_at,
			p.taken_at_local                     AS taken_at_local,
			p.time_zone                          AS time_zone,
			p.photo_title                        AS photo_title,
			p.title_src                          AS title_src,
			p.photo_caption                      AS photo_caption,
			p.caption_src                        AS caption_src,
			p.taken_src                          AS taken_src,
			p.photo_lat                          AS photo_lat,
			p.photo_lng                          AS photo_lng,
			p.place_src                          AS place_src,
			COALESCE(d.keywords, '')             AS details_keywords,
			COALESCE(d.keywords_src, '')         AS keywords_src,
			COALESCE(d.subject, '')              AS details_subject,
			COALESCE(d.subject_src, '')          AS subject_src,
			COALESCE(d.notes, '')                AS details_notes,
			COALESCE(d.notes_src, '')            AS notes_src,
			COALESCE(d.artist, '')               AS details_artist,
			COALESCE(d.artist_src, '')           AS artist_src,
			COALESCE(d.copyright, '')            AS details_copyright,
			COALESCE(d.copyright_src, '')        AS copyright_src
		FROM photos p
		LEFT JOIN details d ON d.photo_id = p.id
		WHERE p.deleted_at IS NULL
		  AND (
			p.title_src       IN ('manual','batch')
			OR p.caption_src  IN ('manual','batch')
			OR p.taken_src    IN ('manual','batch')
			OR p.place_src    IN ('manual','batch')
			OR d.keywords_src IN ('manual','batch')
			OR d.subject_src  IN ('manual','batch')
			OR d.notes_src    IN ('manual','batch')
			OR d.artist_src   IN ('manual','batch')
			OR d.copyright_src IN ('manual','batch')
		  )
		ORDER BY p.updated_at DESC
	`).Scan(&rows).Error; err != nil {
		return fmt.Errorf("dump-edits: query failed: %w", err)
	}

	edits := make([]editRow, 0, len(rows))
	for _, r := range rows {
		fields := map[string]editValue{}

		if isManualSrc(r.TitleSrc) {
			fields["title"] = editValue{Src: r.TitleSrc, Value: r.PhotoTitle}
		}
		if isManualSrc(r.CaptionSrc) {
			fields["caption"] = editValue{Src: r.CaptionSrc, Value: r.PhotoCaption}
		}
		if isManualSrc(r.TakenSrc) {
			fields["taken_at_local"] = editValue{
				Src:   r.TakenSrc,
				Value: takenAtPayload(r.TakenAtLocal, r.TimeZone),
			}
		}
		if isManualSrc(r.PlaceSrc) {
			fields["location"] = editValue{
				Src:   r.PlaceSrc,
				Value: map[string]float64{"lat": r.PhotoLat, "lng": r.PhotoLng},
			}
		}
		if isManualSrc(r.KeywordsSrc) {
			fields["keywords"] = editValue{
				Src:   r.KeywordsSrc,
				Value: hierarchicalKeywords(r.DetailsKeywords),
			}
		}
		if isManualSrc(r.SubjectSrc) {
			fields["subject"] = editValue{Src: r.SubjectSrc, Value: r.DetailsSubject}
		}
		if isManualSrc(r.NotesSrc) {
			fields["notes"] = editValue{Src: r.NotesSrc, Value: r.DetailsNotes}
		}
		if isManualSrc(r.ArtistSrc) {
			fields["artist"] = editValue{Src: r.ArtistSrc, Value: r.DetailsArtist}
		}
		if isManualSrc(r.CopyrightSrc) {
			fields["copyright"] = editValue{Src: r.CopyrightSrc, Value: r.DetailsCopyright}
		}

		if len(fields) == 0 {
			continue
		}

		edits = append(edits, editRow{
			UID:       r.PhotoUID,
			Path:      joinPath(r.PhotoPath, r.PhotoName),
			Original:  r.OriginalName,
			UpdatedAt: r.UpdatedAt,
			Fields:    fields,
		})
	}

	if ctx.Bool("json") {
		return emitJSON(os.Stdout, edits)
	}
	return emitMarkdown(os.Stdout, edits)
}

func isManualSrc(s string) bool {
	return s == string(entity.SrcManual) || s == string(entity.SrcBatch)
}

func joinPath(dir, name string) string {
	if dir == "" {
		return name
	}
	return strings.TrimRight(dir, "/") + "/" + name
}

// hierarchicalKeywords splits PhotoPrism's comma-separated keyword string
// and rewrites each `Parent|Child` segment as `Parent > Child` so the path
// reads naturally in the manifest and matches Lightroom Classic's display
// of hierarchical keywords.
func hierarchicalKeywords(s string) []string {
	out := []string{}
	for _, kw := range strings.Split(s, ",") {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}
		out = append(out, strings.ReplaceAll(kw, "|", " > "))
	}
	sort.Strings(out)
	return out
}

func takenAtPayload(t time.Time, tz string) map[string]string {
	if t.IsZero() {
		return nil
	}
	v := map[string]string{
		"local": t.Format("2006-01-02 15:04:05"),
	}
	if tz != "" && tz != "Local" {
		v["time_zone"] = tz
	}
	return v
}

func emitJSON(w io.Writer, edits []editRow) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(edits)
}

func emitMarkdown(w io.Writer, edits []editRow) error {
	if len(edits) == 0 {
		_, err := fmt.Fprintln(w, "No manually-edited photos found.")
		return err
	}

	fmt.Fprintf(w, "# PhotoPrism manual edits (%d photo%s)\n\n", len(edits), pluralS(len(edits)))
	fmt.Fprintf(w, "_Generated %s_\n\n", time.Now().Format(time.RFC3339))
	fmt.Fprint(w, "Apply each block below in Lightroom (or via an LR-side script), Save\n")
	fmt.Fprint(w, "Metadata to File, then re-index the touched paths with `--overwrite-meta`.\n\n")

	// Stable field ordering inside each photo block.
	fieldOrder := []string{
		"title", "caption", "keywords", "subject", "notes",
		"taken_at_local", "location", "artist", "copyright",
	}

	for _, e := range edits {
		header := e.Path
		if e.Original != "" && e.Original != e.Path {
			header = fmt.Sprintf("%s\n(originally %s)", e.Path, e.Original)
		}
		fmt.Fprintf(w, "## %s\n\n", header)
		fmt.Fprintf(w, "- _UID_: `%s`\n", e.UID)
		fmt.Fprintf(w, "- _Last edit_: %s\n", e.UpdatedAt.Format(time.RFC3339))

		for _, k := range fieldOrder {
			v, ok := e.Fields[k]
			if !ok {
				continue
			}
			fmt.Fprintf(w, "- **%s** _(%s)_: %s\n", k, v.Src, formatValue(v.Value))
		}
		fmt.Fprintln(w)
	}

	return nil
}

func formatValue(v interface{}) string {
	switch x := v.(type) {
	case string:
		return fmt.Sprintf("%q", x)
	case []string:
		if len(x) == 0 {
			return "_(empty)_"
		}
		return strings.Join(x, "; ")
	case map[string]float64:
		return fmt.Sprintf("lat=%.6f, lng=%.6f", x["lat"], x["lng"])
	case map[string]string:
		parts := make([]string, 0, len(x))
		for k, val := range x {
			parts = append(parts, fmt.Sprintf("%s=%s", k, val))
		}
		sort.Strings(parts)
		return strings.Join(parts, ", ")
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

