package designer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"fyne.io/fyne/v2"
)

const draftPrefix = "fynerisor-scene-draft-"

type draftFile struct {
	Ref     string    `json:"ref"`
	Version string    `json:"version"`
	SavedAt time.Time `json:"savedAt"`
	Data    []byte    `json:"data"`
}

func draftName(ref string) string {
	sum := sha256.Sum256([]byte(ref))
	return draftPrefix + hex.EncodeToString(sum[:8]) + ".json"
}

func storage() (fyne.Storage, error) {
	app := fyne.CurrentApp()
	if app == nil || app.Storage() == nil {
		return nil, fmt.Errorf("designer: no app storage")
	}
	return app.Storage(), nil
}

func writeDraft(d draftFile) error {
	st, err := storage()
	if err != nil {
		return err
	}
	w, err := st.Create(draftName(d.Ref))
	if err != nil {
		return err
	}
	defer w.Close()
	enc := json.NewEncoder(w)
	return enc.Encode(d)
}

func readDraft(ref string) (*draftFile, error) {
	st, err := storage()
	if err != nil {
		return nil, err
	}
	r, err := st.Open(draftName(ref))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	data, err := io.ReadAll(io.LimitReader(r, 4<<20))
	if err != nil {
		return nil, err
	}
	var d draftFile
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	if d.Ref != ref {
		return nil, fmt.Errorf("designer: draft belongs to %q", d.Ref)
	}
	return &d, nil
}

func removeDraft(ref string) {
	if st, err := storage(); err == nil {
		_ = st.Remove(draftName(ref))
	}
}
