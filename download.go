package webmod

import (
	"fmt"
	"net/http"
	"path"
)

// DownloadStaticFile downloads a file for the client,
// avoids displaying the file in the browser, forces it to
// directly downloads it instead by setting the content disposition.
// It also allows the specification of the display name
func (t *Tools) DownloadStaticFile(w http.ResponseWriter, r *http.Request, dir, file, displayName string) {
	fp := path.Join(dir, file)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", displayName))
	http.ServeFile(w, r, fp)
}
