package thumb

import (
	"bytes"
	"image"
	"os"

	"github.com/disintegration/imaging"

	"github.com/photoprism/photoprism/pkg/fs"
)

// writeAtomic writes data to fileName atomically. It writes to a sibling
// "<fileName>.tmp" file first and then renames it into place, so concurrent
// HTTP readers cannot observe a partially-written thumbnail. If the rename
// fails the temp file is removed before returning.
func writeAtomic(fileName string, data []byte) error {
	tmp := fileName + ".tmp"

	if err := os.WriteFile(tmp, data, fs.ModeFile); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	if err := os.Rename(tmp, fileName); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	return nil
}

// saveAtomicImage encodes img using the format implied by fileName's extension
// and the given imaging options, then atomically writes it to fileName via
// writeAtomic. This replaces direct imaging.Save calls so concurrent HTTP
// readers never see a half-encoded thumbnail during a reindex.
func saveAtomicImage(img image.Image, fileName string, opts ...imaging.EncodeOption) error {
	format, err := imaging.FormatFromFilename(fileName)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err = imaging.Encode(&buf, img, format, opts...); err != nil {
		return err
	}

	return writeAtomic(fileName, buf.Bytes())
}
