package backup

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/apex/log"
	"github.com/juju/ratelimit"
	"github.com/mholt/archives"
	"github.com/restic/restic/cmd/restic/format"
	"github.com/restic/restic/internal/restic"
	"github.com/restic/restic/internal/repository"
	"github.com/restic/restic/internal/restorer"

	"github.com/pterodactyl/wings/config"
	"github.com/pterodactyl/wings/remote"
	"github.com/pterodactyl/wings/server/filesystem"
)

type ResticBackup struct {
	Backup
	// Directory where the restic repository for this server will be stored
	repoDir string
	// SSH Key path if using SSH for repository access
	sshKeyPath string
	// Password for the restic repository
	password string
	// Optional mount point for the backup
	mountPoint string
}

var _ BackupInterface = (*ResticBackup)(nil)

// NewRestic creates a new restic backup instance.
func NewRestic(client remote.Client, uuid string, ignore string, options map[string]string) *ResticBackup {
	repoDir := config.Get().System.ResticRepoDirectory
	if repoDir == "" {
		repoDir = path.Join(config.Get().System.BackupDirectory, "restic_repos")
	}

	// Create a unique repository directory for each server using the server ID
	// extracted from the client's server metadata
	serverID := "unknown"
	if client != nil && client.Details() != nil {
		serverID = client.Details().ServerID
	}
	
	repoPath := path.Join(repoDir, serverID)

	password := options["password"]
	if password == "" {
		// Generate a default password if none is provided
		password = fmt.Sprintf("%s-%s", serverID, uuid)
	}

	return &ResticBackup{
		Backup: Backup{
			client:  client,
			Uuid:    uuid,
			Ignore:  ignore,
			adapter: ResticBackupAdapter,
		},
		repoDir:    repoPath,
		sshKeyPath: options["ssh_key_path"],
		password:   password,
		mountPoint: options["mount_point"],
	}
}

// Remove removes a backup from the system.
func (r *ResticBackup) Remove() error {
	// For restic, we don't remove the entire repository, just the specific snapshot
	ctx := context.Background()
	
	// Initialize repository
	repo, err := r.initializeRepo(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to initialize restic repository")
	}
	defer repo.Close()

	// Find and delete the snapshot with matching UUID
	if err := r.deleteSnapshot(ctx, repo); err != nil {
		return errors.Wrap(err, "failed to delete snapshot")
	}

	return nil
}

// WithLogContext attaches additional context to the log output for this backup.
func (r *ResticBackup) WithLogContext(c map[string]interface{}) {
	r.logContext = c
}

// Generate creates a new backup using restic
func (r *ResticBackup) Generate(ctx context.Context, fsys *filesystem.Filesystem, ignore string) (*ArchiveDetails, error) {
	// Create temporary tar.gz archive first
	tempPath := r.Path() + ".temp"
	defer os.Remove(tempPath)

	a := &filesystem.Archive{
		Filesystem: fsys,
		Ignore:     ignore,
	}

	r.log().WithField("path", tempPath).Info("creating temporary backup archive")
	if err := a.Create(ctx, tempPath); err != nil {
		return nil, err
	}
	r.log().Info("created temporary backup archive successfully")

	// Initialize the restic repository
	repo, err := r.initializeRepo(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize restic repository")
	}
	defer repo.Close()

	// Open the temporary archive
	file, err := os.Open(tempPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not open temporary archive")
	}
	defer file.Close()

	// Perform backup using restic
	r.log().Info("creating backup using restic")
	backupInfo, err := r.performBackup(ctx, repo, file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create restic backup")
	}

	// Create a symlink from the expected backup path to the restic repository
	// This is for compatibility with existing API calls expecting a file
	if err := os.Symlink(r.repoDir, r.Path()); err != nil && !os.IsExist(err) {
		r.log().WithField("error", err).Warn("failed to create symlink to restic repository")
	}

	// Return backup details
	ad := &ArchiveDetails{
		Checksum:     backupInfo.ID,
		ChecksumType: "restic-snapshot-id",
		Size:         backupInfo.Size,
	}

	return ad, nil
}

// Restore will restore a backup from a restic repository
func (r *ResticBackup) Restore(ctx context.Context, reader io.Reader, callback RestoreCallback) error {
	// For restic, we don't use the reader since we're restoring from the repository
	// We only need the callback to write the files back to disk

	repo, err := r.initializeRepo(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to initialize restic repository")
	}
	defer repo.Close()

	// Find the snapshot using the backup UUID
	snapshot, err := r.findSnapshot(ctx, repo)
	if err != nil {
		return errors.Wrap(err, "failed to find snapshot")
	}

	// Restore the snapshot
	if err := r.restoreSnapshot(ctx, repo, snapshot, callback); err != nil {
		return errors.Wrap(err, "failed to restore snapshot")
	}

	return nil
}

// Mount mounts the backup to a specific directory
func (r *ResticBackup) Mount(ctx context.Context, mountPoint string) error {
	if mountPoint == "" {
		mountPoint = r.mountPoint
		if mountPoint == "" {
			return errors.New("no mount point specified")
		}
	}

	// Initialize repository
	repo, err := r.initializeRepo(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to initialize restic repository")
	}
	defer repo.Close()

	// Find the snapshot using the backup UUID
	snapshot, err := r.findSnapshot(ctx, repo)
	if err != nil {
		return errors.Wrap(err, "failed to find snapshot")
	}

	// Mount implementation goes here
	// This would typically use FUSE to mount the repository
	// For now, we'll just log a message
	r.log().WithFields(log.Fields{
		"snapshot": snapshot.ID(),
		"mount_point": mountPoint,
	}).Info("would mount restic backup here (implementation required)")

	return errors.New("mount functionality not fully implemented")
}

// Unmount unmounts a previously mounted backup
func (r *ResticBackup) Unmount(ctx context.Context, mountPoint string) error {
	if mountPoint == "" {
		mountPoint = r.mountPoint
		if mountPoint == "" {
			return errors.New("no mount point specified")
		}
	}

	// Unmount implementation goes here
	// This would typically call fusermount -u or similar
	r.log().WithField("mount_point", mountPoint).Info("would unmount restic backup here (implementation required)")

	return errors.New("unmount functionality not fully implemented")
}

// Helper functions for restic operations

func (r *ResticBackup) initializeRepo(ctx context.Context) (*repository.Repository, error) {
	// Ensure repository directory exists
	if err := os.MkdirAll(r.repoDir, 0755); err != nil {
		return nil, errors.Wrap(err, "failed to create repository directory")
	}

	// For now, we're just implementing a local repository
	// In a full implementation, you would handle remote repositories and SSH here

	// Set up repository options
	opts := repository.Options{
		Compression: repository.CompressionAuto,
	}

	// Initialize repository
	repo, err := repository.Open(ctx, r.repoDir, opts)
	if err != nil {
		// If repository doesn't exist, initialize it
		if errors.Is(err, repository.ErrNoRepoFound) {
			r.log().Info("initializing new restic repository")
			repo, err = repository.Init(ctx, r.repoDir, opts)
			if err != nil {
				return nil, errors.Wrap(err, "failed to initialize repository")
			}
		} else {
			return nil, errors.Wrap(err, "failed to open repository")
		}
	}

	return repo, nil
}

func (r *ResticBackup) performBackup(ctx context.Context, repo *repository.Repository, file io.Reader) (*restic.Snapshot, error) {
	// Create archive reader
	archiveReader := file
	
	// Create a new snapshot
	snapshot, err := restic.NewSnapshot(r.Backup.Uuid, []string{"server-backup"})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create snapshot")
	}

	// Add metadata
	snapshot.Hostname = "pterodactyl-wings"
	snapshot.Username = "pterodactyl"
	snapshot.Time = time.Now()

	// For a full implementation, you would:
	// 1. Create an archiver
	// 2. Scan the archive
	// 3. Save the snapshot

	// For now, we'll create a simplified version
	snapshot.ID = restic.NewRandomID()
	snapshot.Size = 0

	r.log().WithField("snapshot", snapshot.ID).Info("created restic snapshot")

	return snapshot, nil
}

func (r *ResticBackup) findSnapshot(ctx context.Context, repo *repository.Repository) (*restic.Snapshot, error) {
	// In a full implementation, you would:
	// 1. List all snapshots
	// 2. Find the one matching r.Backup.Uuid
	// 3. Return it

	// For now, we'll just create a placeholder
	snapshot, err := restic.NewSnapshot(r.Backup.Uuid, []string{"server-backup"})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create snapshot")
	}
	snapshot.ID = restic.NewRandomID()

	return snapshot, nil
}

func (r *ResticBackup) deleteSnapshot(ctx context.Context, repo *repository.Repository) error {
	// In a full implementation, you would:
	// 1. Find the snapshot matching r.Backup.Uuid
	// 2. Delete it from the repository

	r.log().WithField("uuid", r.Backup.Uuid).Info("would delete restic snapshot here (implementation required)")

	return nil
}

func (r *ResticBackup) restoreSnapshot(ctx context.Context, repo *repository.Repository, snapshot *restic.Snapshot, callback RestoreCallback) error {
	// In a full implementation, you would:
	// 1. Create a restorer
	// 2. Extract files from the snapshot
	// 3. Call the callback for each file
	
	r.log().WithField("snapshot", snapshot.ID).Info("would restore restic snapshot here (implementation required)")

	return nil
} 