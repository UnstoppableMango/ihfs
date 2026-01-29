package cowfs

import "github.com/unstoppablemango/ihfs"

type MergeStrategy func(layer, base []ihfs.DirEntry) ([]ihfs.DirEntry, error)

var DefaultMergeStrategy MergeStrategy = mergeDirEntries

// mergeDirEntries merges directory entries from the layer and base,
// with layer entries taking precedence over base entries with the same name.
func mergeDirEntries(layer, base []ihfs.DirEntry) ([]ihfs.DirEntry, error) {
	entries := make(map[string]ihfs.DirEntry)

	for _, entry := range layer {
		entries[entry.Name()] = entry
	}

	for _, entry := range base {
		if _, exists := entries[entry.Name()]; !exists {
			entries[entry.Name()] = entry
		}
	}

	result := make([]ihfs.DirEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry)
	}

	return result, nil
}
