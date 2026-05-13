package upload

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/tus/tusd/v2/pkg/filestore"
	tusd "github.com/tus/tusd/v2/pkg/handler"

	"github.com/JMC50/nas/internal/auth"
	"github.com/JMC50/nas/internal/db"
	"github.com/JMC50/nas/internal/files"
)

// MountTus creates and registers a tusd handler at /files/ on the chi router.
// Per-upload metadata fields:
//
//	filename — required, becomes the basename in nas-data
//	loc      — optional subdirectory under NAS_DATA_DIR (default "")
//
// On successful upload completion (CompleteUploads event), the staged file is
// moved to NAS_DATA_DIR/<loc>/<filename> and an activity log is written.
//
// JWT verification happens via the chi middleware ahead of /files/; tusd itself
// runs un-authenticated here. We can tighten with PreUploadCreateCallback later.
func (h *Handlers) MountTus(stagingDir string) (http.Handler, error) {
	if err := os.MkdirAll(stagingDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir staging: %w", err)
	}
	store := filestore.New(stagingDir)
	composer := tusd.NewStoreComposer()
	store.UseIn(composer)

	handler, err := tusd.NewHandler(tusd.Config{
		BasePath:                "/files/",
		StoreComposer:           composer,
		NotifyCompleteUploads:   true,
		RespectForwardedHeaders: true,
		PreUploadCreateCallback: h.preCreateHook,
	})
	if err != nil {
		return nil, fmt.Errorf("new tusd handler: %w", err)
	}

	go h.watchCompletions(handler)

	return handler, nil
}

// preCreateHook validates that the upload has a JWT (via Authorization header,
// since tus clients sometimes can't pass query params). Verifies UPLOAD intent.
func (h *Handlers) preCreateHook(hook tusd.HookEvent) (tusd.HTTPResponse, tusd.FileInfoChanges, error) {
	authHeader := hook.HTTPRequest.Header.Get("Authorization")
	if len(authHeader) <= len("Bearer ") || authHeader[:7] != "Bearer " {
		return tusd.HTTPResponse{}, tusd.FileInfoChanges{},
			tusd.NewError("ERR_UNAUTHORIZED", "token required", http.StatusUnauthorized)
	}
	token := authHeader[7:]
	claims, err := auth.ParseToken(token, h.Config.PrivateKey)
	if err != nil {
		return tusd.HTTPResponse{}, tusd.FileInfoChanges{},
			tusd.NewError("ERR_UNAUTHORIZED", "invalid token", http.StatusUnauthorized)
	}
	allowed, err := db.HasIntent(h.DB, claims.UserID, "UPLOAD")
	if err != nil {
		return tusd.HTTPResponse{}, tusd.FileInfoChanges{},
			tusd.NewError("ERR_INTERNAL", "intent check failed", http.StatusInternalServerError)
	}
	if !allowed {
		return tusd.HTTPResponse{}, tusd.FileInfoChanges{},
			tusd.NewError("ERR_FORBIDDEN", "no UPLOAD intent", http.StatusForbidden)
	}
	// Stash userId for the post-completion hook
	metaChange := tusd.FileInfoChanges{}
	if hook.Upload.MetaData != nil {
		metaChange.MetaData = map[string]string{}
		for key, value := range hook.Upload.MetaData {
			metaChange.MetaData[key] = value
		}
		metaChange.MetaData["userId"] = claims.UserID
	}
	return tusd.HTTPResponse{}, metaChange, nil
}

// watchCompletions listens on the tusd CompleteUploads channel and moves
// staged files to their final NAS_DATA_DIR location.
func (h *Handlers) watchCompletions(handler *tusd.Handler) {
	for event := range handler.CompleteUploads {
		if err := h.finalizeUpload(event); err != nil {
			slog.Error("finalize upload failed", "err", err, "id", event.Upload.ID)
		}
	}
}

func (h *Handlers) finalizeUpload(event tusd.HookEvent) error {
	filename := filepath.Base(event.Upload.MetaData["filename"])
	if filename == "" || filename == "." || filename == string(filepath.Separator) {
		return errors.New("upload missing filename metadata")
	}
	loc := event.Upload.MetaData["loc"]
	userID := event.Upload.MetaData["userId"]

	target, err := files.SafeJoin(h.Config.NASDataDir, files.TrimLeadingSlash(loc), filename)
	if err != nil {
		return fmt.Errorf("safe join: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("mkdir target: %w", err)
	}

	stagedPath := filepath.Join(h.Config.NASTempDir, "tus", event.Upload.ID)
	if err := os.Rename(stagedPath, target); err != nil {
		// Cross-FS rename may fail — try copy+remove
		if err := copyFileBytes(stagedPath, target); err != nil {
			return fmt.Errorf("move staged: %w", err)
		}
		_ = os.Remove(stagedPath)
	}
	_ = os.Remove(stagedPath + ".info")

	if userID != "" {
		_ = db.InsertLog(h.DB, userID, "UPLOAD",
			fmt.Sprintf("UPLOAD [FILE] AT /%s", convertLoc(loc+"/"+filename)),
			"/"+loc, time.Now().UnixMilli())
	}
	return nil
}

func copyFileBytes(source, target string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(target)
	if err != nil {
		return err
	}
	defer out.Close()
	buffer := make([]byte, 64*1024)
	for {
		n, err := in.Read(buffer)
		if n > 0 {
			if _, werr := out.Write(buffer[:n]); werr != nil {
				return werr
			}
		}
		if err != nil {
			if err.Error() == "EOF" {
				return nil
			}
			return err
		}
	}
}
