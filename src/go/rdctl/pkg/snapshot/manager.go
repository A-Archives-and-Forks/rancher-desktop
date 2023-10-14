package snapshot

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/rancher-sandbox/rancher-desktop/src/go/rdctl/pkg/paths"
)

var nameRegexp = *regexp.MustCompile("^[0-9a-zA-Z_-]{0,100}$")
var ErrNameExists = errors.New("name already exists")
var ErrInvalidName = fmt.Errorf("name does not match regex %q", nameRegexp.String())

// Writes the data in a Snapshot to the metadata.json file in a snapshot
// directory. This is done last because we consider the presence of this file to
// be the hallmark of a complete and valid snapshot.
func writeMetadataFile(paths paths.Paths, snapshot Snapshot) error {
	metadataPath := filepath.Join(paths.Snapshots, snapshot.ID, "metadata.json")
	metadataFile, err := os.Create(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer metadataFile.Close()
	encoder := json.NewEncoder(metadataFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(snapshot); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}
	return nil
}

// Handles all snapshot-related functionality.
type Manager struct {
	Paths       paths.Paths
	Snapshotter Snapshotter
}

func NewManager(paths paths.Paths) Manager {
	return Manager{
		Paths:       paths,
		Snapshotter: NewSnapshotterImpl(paths),
	}
}

func (manager *Manager) GetSnapshotId(desiredName string) (string, error) {
	snapshots, err := manager.List()
	if err != nil {
		return "", fmt.Errorf("failed to list snapshots: %w", err)
	}
	for _, candidate := range snapshots {
		if desiredName == candidate.Name {
			return candidate.ID, nil
		}
	}
	return "", fmt.Errorf(`can't find snapshot %q`, desiredName)
}

// Creates a new snapshot.
func (manager Manager) Create(name string) (*Snapshot, error) {
	// validate name
	currentSnapshots, err := manager.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}
	for _, currentSnapshot := range currentSnapshots {
		if currentSnapshot.Name == name {
			return nil, fmt.Errorf("invalid name %q: %w", name, ErrNameExists)
		}
	}
	if !nameRegexp.MatchString(name) {
		return nil, fmt.Errorf("invalid name %q: %w", name, ErrInvalidName)
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID for snapshot: %w", err)
	}
	snapshot := Snapshot{
		Created: time.Now(),
		Name:    name,
		ID:      id.String(),
	}

	// do operations that can fail, rolling back if failure is encountered
	snapshotDir := filepath.Join(manager.Paths.Snapshots, snapshot.ID)
	if err := manager.Snapshotter.CreateFiles(snapshot); err != nil {
		if err := os.RemoveAll(snapshotDir); err != nil {
			return nil, fmt.Errorf("failed to delete created snapshot directory: %w", err)
		}
		return nil, fmt.Errorf("failed to consummate snapshot: %w", err)
	}

	return &snapshot, nil
}

// Returns snapshots that are present on system.
func (manager Manager) List() ([]Snapshot, error) {
	dirEntries, err := os.ReadDir(manager.Paths.Snapshots)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return []Snapshot{}, fmt.Errorf("failed to read snapshots directory: %w", err)
	}
	snapshots := make([]Snapshot, 0, len(dirEntries))
	for _, dirEntry := range dirEntries {
		snapshot := Snapshot{}
		metadataPath := filepath.Join(manager.Paths.Snapshots, dirEntry.Name(), "metadata.json")
		contents, err := os.ReadFile(metadataPath)
		if err != nil {
			return []Snapshot{}, fmt.Errorf("failed to read %q: %w", metadataPath, err)
		}
		if err := json.Unmarshal(contents, &snapshot); err != nil {
			return []Snapshot{}, fmt.Errorf("failed to unmarshal contents of %q: %w", metadataPath, err)
		}
		snapshot.Created = snapshot.Created.Local()
		snapshots = append(snapshots, snapshot)
	}
	return snapshots, nil
}

// Deletes a snapshot.
func (manager Manager) Delete(id string) error {
	dirEntries, err := os.ReadDir(manager.Paths.Snapshots)
	if err != nil {
		return fmt.Errorf("failed to read snapshots dir: %w", err)
	}
	found := false
	for _, dirEntry := range dirEntries {
		if dirEntry.Name() == id {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("snapshot with id %q does not exist", id)
	}
	snapshotDir := filepath.Join(manager.Paths.Snapshots, id)
	if err = os.RemoveAll(snapshotDir); err != nil {
		return fmt.Errorf("failed to remove dir %q: %w", snapshotDir, err)
	}
	return nil
}

// Restores Rancher Desktop to the state saved in a snapshot.
func (manager Manager) Restore(id string) error {
	metadataPath := filepath.Join(manager.Paths.Snapshots, id, "metadata.json")
	contents, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read metadata for snapshot %q: %w", id, err)
	}
	snapshot := Snapshot{}
	if err := json.Unmarshal(contents, &snapshot); err != nil {
		return fmt.Errorf("failed to unmarshal contents of %q: %w", metadataPath, err)
	}

	if err := manager.Snapshotter.RestoreFiles(snapshot); err != nil {
		return fmt.Errorf("failed to restore files: %w", err)
	}

	return nil
}
