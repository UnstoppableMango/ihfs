package union

import "github.com/unstoppablemango/ihfs"

type MergeStrategy func(layer, base []ihfs.DirEntry) ([]ihfs.DirEntry, error)

var DefaultMergeStrategy MergeStrategy = mergeDirEntries

// mergeDirEntries merges directory entries from the layer and base,
// with layer entries taking precedence over base entries with the same name.
// The order is maintained by first including all layer entries, then base entries
// that don't exist in the layer.
func mergeDirEntries(layer, base []ihfs.DirEntry) ([]ihfs.DirEntry, error) {
	seen := make(map[string]bool)
	result := make([]ihfs.DirEntry, 0, len(layer)+len(base))

	// Add all layer entries first
	for _, entry := range layer {
		result = append(result, entry)
		seen[entry.Name()] = true
	}

	// Add base entries that don't exist in layer
	for _, entry := range base {
		if !seen[entry.Name()] {
			result = append(result, entry)
		}
	}

	return result, nil
}
