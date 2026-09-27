package main

import (
	"bytes"
	"context"
	"crypto/md5" //nolint:gosec // MD5 is used only for content diffing, not security.
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/disintegration/imaging"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

//////////////////////////////////////////////////////////////////////////////
//
// sync-media command
//
// Uploads all media (images, PDFs, videos, docs, etc.) under
// content/{articles,pages,images} to the configured Azure Blob container,
// mirroring the repo layout in blob keys. Images are optimized (max 1200x1200,
// JPEG quality 85) before upload using the same rules as the old local build
// step. Blobs are never removed from the container; delete them manually if
// you need to reclaim space.
//
//////////////////////////////////////////////////////////////////////////////

const (
	defaultStorageAccount = "coolstercodes"
	defaultContainerName  = "public"
	defaultConcurrency    = 8
)

// mediaExtensions is the set of file extensions treated as "media" (i.e.
// non-code) that will be synced to blob storage. Keep in lowercase.
var mediaExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".svg":  true,
	".webp": true,
	".ico":  true,
	".pdf":  true,
	".mp4":  true,
	".mov":  true,
	".webm": true,
	".mp3":  true,
	".wav":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
	".ppt":  true,
	".pptx": true,
	".zip":  true,
	".MOV":  true,
}

// syncMediaPrefixes are the repo-relative directory roots we sync from.
var syncMediaPrefixes = []string{
	"content/articles",
	"content/pages",
	"content/images",
}

type syncMediaFlags struct {
	account     string
	containerNm string
	dryRun      bool
	concurrency int
	watch       bool
	debounceMs  int
}

func newSyncMediaCommand() *cobra.Command {
	f := &syncMediaFlags{
		account:     defaultStorageAccount,
		containerNm: defaultContainerName,
		concurrency: defaultConcurrency,
		debounceMs:  750,
	}

	cmd := &cobra.Command{
		Use:   "sync-media",
		Short: "Sync local media (images, PDFs, videos, docs) to Azure Blob Storage",
		Long: strings.TrimSpace(`
Walks content/{articles,pages,images}, optimizes images, and uploads every
media file to the configured Azure Blob Storage container, mirroring the repo
layout. Blobs are never removed from the container; delete them manually if
you need to reclaim space. Authenticates via DefaultAzureCredential (typically
'az login').

With --watch, performs an initial sync and then keeps running, re-syncing
whenever a media file is added or changed under the media roots.
`),
		RunE: func(_ *cobra.Command, _ []string) error {
			return runSyncMedia(context.Background(), f)
		},
	}

	cmd.Flags().StringVar(&f.account, "account", f.account, "Azure Storage account name")
	cmd.Flags().StringVar(&f.containerNm, "container", f.containerNm, "Azure Blob container name")
	cmd.Flags().BoolVar(&f.dryRun, "dry-run", false, "Log actions without uploading")
	cmd.Flags().IntVar(&f.concurrency, "concurrency", f.concurrency, "Number of parallel upload workers")
	cmd.Flags().BoolVar(&f.watch, "watch", false, "Keep running and re-sync on filesystem changes")
	cmd.Flags().IntVar(&f.debounceMs, "debounce-ms", f.debounceMs, "Debounce window for coalescing filesystem events (ms)")

	return cmd
}

func runSyncMedia(ctx context.Context, f *syncMediaFlags) error {
	log := getLog()

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return xerrors.Errorf("acquiring Azure credentials (is 'az login' active?): %w", err)
	}

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", f.account)
	client, err := azblob.NewClient(serviceURL, cred, nil)
	if err != nil {
		return xerrors.Errorf("creating blob client: %w", err)
	}
	containerClient := client.ServiceClient().NewContainerClient(f.containerNm)

	if err := syncOnce(ctx, log, containerClient, f); err != nil {
		if !f.watch {
			return err
		}
		log.Errorf("Initial sync failed (continuing in watch mode): %v", err)
	}

	if !f.watch {
		return nil
	}
	return watchAndSync(ctx, log, containerClient, f)
}

// syncOnce performs a single upload+delete pass.
func syncOnce(
	ctx context.Context,
	log *logrus.Logger,
	containerClient *container.Client,
	f *syncMediaFlags,
) error {
	local, err := collectLocalMedia(log)
	if err != nil {
		return err
	}
	log.Infof("Discovered %d local media files across %d prefixes", len(local), len(syncMediaPrefixes))

	uploaded, skipped, err := uploadAll(ctx, log, containerClient, local, f)
	if err != nil {
		return err
	}
	log.Infof("Upload phase complete: uploaded=%d skipped=%d", uploaded, skipped)

	log.Infof("Sync complete: uploaded=%d skipped=%d dryRun=%v", uploaded, skipped, f.dryRun)
	return nil
}

// watchAndSync watches the media prefixes for filesystem changes and triggers
// a debounced syncOnce whenever media files are created, modified, or removed.
func watchAndSync(
	ctx context.Context,
	log *logrus.Logger,
	containerClient *container.Client,
	f *syncMediaFlags,
) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return xerrors.Errorf("creating fs watcher: %w", err)
	}
	defer w.Close()

	// Add every existing directory under each media prefix. fsnotify is not
	// recursive on macOS/Linux, so we also add new directories as they appear.
	addDirsRecursively := func(root string) {
		_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err != nil || d == nil {
				return nil //nolint:nilerr // skip unreadable entries and keep walking
			}
			if d.IsDir() {
				if addErr := w.Add(p); addErr != nil {
					log.Debugf("watcher: failed to add %s: %v", p, addErr)
				}
			}
			return nil
		})
	}
	for _, prefix := range syncMediaPrefixes {
		if _, err := os.Stat(prefix); err == nil {
			addDirsRecursively(prefix)
		}
	}

	log.Infof("Watching for media changes under %v (debounce=%dms). Ctrl-C to stop.",
		syncMediaPrefixes, f.debounceMs)

	debounce := time.NewTimer(time.Hour)
	debounce.Stop()
	pending := false

	for {
		select {
		case <-ctx.Done():
			return nil

		case ev, ok := <-w.Events:
			if !ok {
				return nil
			}
			// If a new subdirectory was created, watch it too.
			if ev.Op&fsnotify.Create != 0 {
				if info, statErr := os.Stat(ev.Name); statErr == nil && info.IsDir() {
					addDirsRecursively(ev.Name)
				}
			}
			// Only trigger for media file events (or dir events that touched media).
			if !isMediaPath(ev.Name) && !isLikelyDirEvent(ev) {
				continue
			}
			log.Debugf("watcher event: %s %s", ev.Op, ev.Name)
			if !pending {
				pending = true
			}
			debounce.Reset(time.Duration(f.debounceMs) * time.Millisecond)

		case err, ok := <-w.Errors:
			if !ok {
				return nil
			}
			log.Errorf("watcher error: %v", err)

		case <-debounce.C:
			pending = false
			log.Info("Media change detected; syncing...")
			if err := syncOnce(ctx, log, containerClient, f); err != nil {
				log.Errorf("Sync failed: %v", err)
			}
		}
	}
}

func isMediaPath(p string) bool {
	return mediaExtensions[strings.ToLower(filepath.Ext(p))]
}

// isLikelyDirEvent returns true if the event has no file extension (usually a
// directory create/rename/remove that we still want to react to so newly
// dropped folders trigger a sync).
func isLikelyDirEvent(ev fsnotify.Event) bool {
	return filepath.Ext(ev.Name) == ""
}

// localMediaFile is a single file we intend to upload. Content is the bytes to
// upload (post-optimization for images) and MD5 is the checksum of those bytes.
type localMediaFile struct {
	SourcePath  string // absolute or repo-relative path on disk
	BlobKey     string // repo-relative, unix-style; matches blob path in container
	Content     []byte
	MD5         []byte
	ContentType string
}

func collectLocalMedia(log *logrus.Logger) (map[string]*localMediaFile, error) {
	out := make(map[string]*localMediaFile)
	for _, prefix := range syncMediaPrefixes {
		if _, err := os.Stat(prefix); errors.Is(err, os.ErrNotExist) {
			log.Debugf("Skipping missing prefix: %s", prefix)
			continue
		}
		err := filepath.WalkDir(prefix, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(p))
			if !mediaExtensions[ext] {
				return nil
			}
			key := filepath.ToSlash(p)
			out[key] = &localMediaFile{
				SourcePath: p,
				BlobKey:    key,
			}
			return nil
		})
		if err != nil {
			return nil, xerrors.Errorf("walking %s: %w", prefix, err)
		}
	}
	return out, nil
}

func uploadAll(
	ctx context.Context,
	log *logrus.Logger,
	containerClient *container.Client,
	local map[string]*localMediaFile,
	f *syncMediaFlags,
) (int, int, error) {
	type result struct {
		key      string
		uploaded bool
		err      error
	}

	jobs := make(chan *localMediaFile)
	results := make(chan result)

	var wg sync.WaitGroup
	for range f.concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for m := range jobs {
				didUpload, uerr := uploadOne(ctx, log, containerClient, m, f)
				results <- result{key: m.BlobKey, uploaded: didUpload, err: uerr}
			}
		}()
	}

	go func() {
		for _, m := range local {
			jobs <- m
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	var (
		uploaded int
		skipped  int
		firstErr error
	)
	for r := range results {
		if r.err != nil {
			log.Errorf("Upload failed for %s: %v", r.key, r.err)
			if firstErr == nil {
				firstErr = r.err
			}
			continue
		}
		if r.uploaded {
			uploaded++
		} else {
			skipped++
		}
	}
	return uploaded, skipped, firstErr
}

func uploadOne(
	ctx context.Context,
	log *logrus.Logger,
	containerClient *container.Client,
	m *localMediaFile,
	f *syncMediaFlags,
) (bool, error) {
	// Prepare content (optimize images; read raw for everything else).
	if err := prepareContent(log, m); err != nil {
		return false, err
	}

	blobClient := containerClient.NewBlobClient(m.BlobKey)
	if same, err := blobMatches(ctx, blobClient, m.MD5); err != nil {
		return false, err
	} else if same {
		log.Debugf("Unchanged, skipping: %s", m.BlobKey)
		return false, nil
	}

	if f.dryRun {
		log.Infof("[dry-run] Would upload %s (%d bytes)", m.BlobKey, len(m.Content))
		return true, nil
	}

	blockBlob := containerClient.NewBlockBlobClient(m.BlobKey)
	_, err := blockBlob.UploadBuffer(ctx, m.Content, &azblob.UploadBufferOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: to.Ptr(m.ContentType),
			BlobContentMD5:  m.MD5,
		},
	})
	if err != nil {
		return false, xerrors.Errorf("uploading %s: %w", m.BlobKey, err)
	}
	log.Infof("Uploaded %s (%d bytes)", m.BlobKey, len(m.Content))
	return true, nil
}

// prepareContent reads the source file and, for JPEG/PNG images that exceed
// MaxImageWidth/MaxImageHeight, resizes them using the same rules as the old
// local mphoto step. The optimized bytes are stored on the localMediaFile.
func prepareContent(log *logrus.Logger, m *localMediaFile) error {
	raw, err := os.ReadFile(m.SourcePath)
	if err != nil {
		return xerrors.Errorf("reading %s: %w", m.SourcePath, err)
	}

	ext := strings.ToLower(filepath.Ext(m.SourcePath))
	m.ContentType = contentTypeForExt(ext)

	optimized, ok, err := maybeOptimizeImage(log, m.SourcePath, ext, raw)
	if err != nil {
		return err
	}
	if ok {
		m.Content = optimized
	} else {
		m.Content = raw
	}

	sum := md5.Sum(m.Content) //nolint:gosec
	m.MD5 = sum[:]
	return nil
}

func maybeOptimizeImage(log *logrus.Logger, path, ext string, raw []byte) ([]byte, bool, error) {
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return nil, false, nil
	}
	img, err := imaging.Decode(bytes.NewReader(raw), imaging.AutoOrientation(true))
	if err != nil {
		// Non-decodable image — upload as-is.
		log.Debugf("Could not decode image %s, uploading as-is: %v", path, err)
		return nil, false, nil
	}
	b := img.Bounds()
	if b.Dx() <= MaxImageWidth && b.Dy() <= MaxImageHeight {
		return nil, false, nil
	}
	newW, newH := calculateOptimizedDims(b.Dx(), b.Dy(), MaxImageWidth, MaxImageHeight)
	log.Infof("Optimizing %s from %dx%d to %dx%d", path, b.Dx(), b.Dy(), newW, newH)
	resized := imaging.Resize(img, newW, newH, imaging.Lanczos)

	var buf bytes.Buffer
	switch ext {
	case ".jpg", ".jpeg":
		if err := imaging.Encode(&buf, resized, imaging.JPEG, imaging.JPEGQuality(ImageQuality)); err != nil {
			return nil, false, xerrors.Errorf("encoding JPEG %s: %w", path, err)
		}
	case ".png":
		if err := imaging.Encode(&buf, resized, imaging.PNG); err != nil {
			return nil, false, xerrors.Errorf("encoding PNG %s: %w", path, err)
		}
	}
	return buf.Bytes(), true, nil
}

// calculateOptimizedDims scales (w,h) to fit inside (maxW,maxH) preserving
// aspect ratio. It mirrors mphoto.calculateNewDimensions.
func calculateOptimizedDims(w, h, maxW, maxH int) (int, int) {
	if w <= maxW && h <= maxH {
		return w, h
	}
	ratioW := float64(maxW) / float64(w)
	ratioH := float64(maxH) / float64(h)
	ratio := ratioW
	if ratioH < ratio {
		ratio = ratioH
	}
	return int(float64(w) * ratio), int(float64(h) * ratio)
}

// blobMatches returns true if the remote blob's MD5 matches localMD5.
func blobMatches(ctx context.Context, bc *blob.Client, localMD5 []byte) (bool, error) {
	props, err := bc.GetProperties(ctx, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.ErrorCode == string(bloberror.BlobNotFound) {
			return false, nil
		}
		return false, xerrors.Errorf("HEAD blob: %w", err)
	}
	if len(props.ContentMD5) == 0 {
		return false, nil
	}
	return bytes.Equal(props.ContentMD5, localMD5), nil
}

func contentTypeForExt(ext string) string {
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	case ".ico":
		return "image/x-icon"
	case ".pdf":
		return "application/pdf"
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	case ".webm":
		return "video/webm"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}

// Ensure io is referenced (kept for future streaming uploads).
var _ = io.Discard
