package thumb

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/disintegration/imaging"
	"github.com/stretchr/testify/assert"
)

func TestWriteAtomic(t *testing.T) {
	t.Run("WritesAndRenames", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "thumb.jpg")
		payload := []byte("hello world")

		if err := writeAtomic(target, payload); err != nil {
			t.Fatal(err)
		}

		got, err := os.ReadFile(target)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, payload, got)

		// The .tmp file should not linger after a successful write.
		_, err = os.Stat(target + ".tmp")
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("OverwritesExistingFile", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "thumb.jpg")
		assert.NoError(t, os.WriteFile(target, []byte("old"), 0o644))
		assert.NoError(t, writeAtomic(target, []byte("new")))

		got, err := os.ReadFile(target)
		assert.NoError(t, err)
		assert.Equal(t, []byte("new"), got)
	})
}

func TestSaveAtomicImage(t *testing.T) {
	t.Run("RoundTripsAJpeg", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "thumb.jpg")

		img := image.NewRGBA(image.Rect(0, 0, 4, 4))
		img.Set(0, 0, color.RGBA{R: 255, A: 255})

		if err := saveAtomicImage(img, target, imaging.JPEGQuality(80)); err != nil {
			t.Fatal(err)
		}

		// Re-open and verify dimensions.
		decoded, err := imaging.Open(target)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 4, decoded.Bounds().Dx())
		assert.Equal(t, 4, decoded.Bounds().Dy())

		// No leftover temp file.
		_, err = os.Stat(target + ".tmp")
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("RejectsUnknownExtension", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "thumb.xyz")
		img := image.NewRGBA(image.Rect(0, 0, 2, 2))

		err := saveAtomicImage(img, target)
		assert.Error(t, err)
	})
}
