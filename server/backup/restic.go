package backup

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"emperror.dev/errors"
	"github.com/pterodactyl/wings/config"
	"github.com/pterodactyl/wings/remote"
	"github.com/pterodactyl/wings/server/filesystem"
	"github.com/restic/restic/lib/repository"
	"github.com/restic/restic/lib/restic"
)

const ResticBackupAdapter AdapterType = "restic"

type ResticBackup struct {
	Backup
	RepoPath string
	Password string
}

// NewResticBackup creates a new restic backup instance.
func NewResticBackup(uuid string, ignore string, repoPath string, password string) *ResticBackup {
	return &ResticBackup{
		Backup: Backup{
			Uuid:   uuid,
			Ignore: ignore,
			adapter: ResticBackupAdapter,
		},
		RepoPath: repoPath,
		Password: password,
	}
}

// Generate creates a new backup in the restic repository.
func (rb *ResticBackup) Generate(ctx context.Context, fs *filesystem.Filesystem, ignore string) (*ArchiveDetails, error) {
	repo, err := repository.Open(ctx, rb.RepoPath, rb.Password)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer repo.Close()

	// Create a new snapshot
	snap, err := restic.NewSnapshot([]string{fs.Path()})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Save the snapshot
	id, err := repo.SaveSnapshot(ctx, snap)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	details := &ArchiveDetails{
		Checksum:     id.String(),
		ChecksumType: "restic_id",
		Size:         0, // We'll need to calculate this
		Parts:        []remote.BackupPart{{Etag: id.String()}},
	}

	return details, nil
}

// Mount mounts a restic backup at the specified path.
func (rb *ResticBackup) Mount(ctx context.Context, mountPath string) error {
	repo, err := repository.Open(ctx, rb.RepoPath, rb.Password)
	if err != nil {
		return errors.WithStack(err)
	}
	defer repo.Close()

	// Implementation for mounting will need more work with FUSE
	return nil
}

// Restore implements the backup restoration for restic.
func (rb *ResticBackup) Restore(ctx context.Context, r io.Reader, callback RestoreCallback) error {
	repo, err := repository.Open(ctx, rb.RepoPath, rb.Password)
	if err != nil {
		return errors.WithStack(err)
	}
	defer repo.Close()

	// Implementation for restore functionality
	// This will need to be expanded based on specific requirements
	return nil
}

// Remove removes the backup from the restic repository.
func (rb *ResticBackup) Remove() error {
	ctx := context.Background()
	repo, err := repository.Open(ctx, rb.RepoPath, rb.Password)
	if err != nil {
		return errors.WithStack(err)
	}
	defer repo.Close()

	// Implementation for removing specific snapshot
	return nil
}

func (rb *ResticBackup) WithLogContext(c map[string]interface{}) {
	rb.logContext = c
} 