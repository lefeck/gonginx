package utils

import (
	"fmt"
	"sort"
	"strings"

	"github.com/lefeck/gonginx/config"
)

// DiffType represents the type of difference
type DiffType int

const (
	// DiffAdded represents an added directive
	DiffAdded DiffType = iota
	// DiffRemoved represents a removed directive
	DiffRemoved
	// DiffModified represents a modified directive
	DiffModified
	// DiffMoved represents a moved directive
	DiffMoved
)

// String returns the string representation of the diff type
func (dt DiffType) String() string {
	switch dt {
	case DiffAdded:
		return "ADDED"
	case DiffRemoved:
		return "REMOVED"
	case DiffModified:
		return "MODIFIED"
	case DiffMoved:
		return "MOVED"
	default:
		return "UNKNOWN"
	}
}

// Difference represents a single difference between configurations
type Difference struct {
	Type          DiffType
	Path          string
	DirectiveName string
	OldValue      string
	NewValue      string
	Context       string
	Line          int
	Description   string
}

// String returns a human-readable representation of the difference
func (d *Difference) String() string {
	switch d.Type {
	case DiffAdded:
		return fmt.Sprintf("+ [%s] %s: %s", d.Path, d.DirectiveName, d.NewValue)
	case DiffRemoved:
		return fmt.Sprintf("- [%s] %s: %s", d.Path, d.DirectiveName, d.OldValue)
	case DiffModified:
		return fmt.Sprintf("~ [%s] %s: %s -> %s", d.Path, d.DirectiveName, d.OldValue, d.NewValue)
	case DiffMoved:
		return fmt.Sprintf("^ [%s] %s: moved to %s", d.Path, d.DirectiveName, d.NewValue)
	default:
		return fmt.Sprintf("? [%s] %s", d.Path, d.DirectiveName)
	}
}

// DiffResult represents the result of a configuration comparison
type DiffResult struct {
	Differences []Difference
	Summary     DiffSummary
}

// DiffSummary provides summary statistics about the differences
type DiffSummary struct {
	Total    int
	Added    int
	Removed  int
	Modified int
	Moved    int
}

// String returns a summary of the differences
func (ds *DiffSummary) String() string {
	return fmt.Sprintf("Total: %d, Added: %d, Removed: %d, Modified: %d, Moved: %d",
		ds.Total, ds.Added, ds.Removed, ds.Modified, ds.Moved)
}

// HasChanges returns true if there are any differences
func (dr *DiffResult) HasChanges() bool {
	return len(dr.Differences) > 0
}

// GetByType returns differences of a specific type
func (dr *DiffResult) GetByType(diffType DiffType) []Difference {
	var filtered []Difference
	for _, diff := range dr.Differences {
		if diff.Type == diffType {
			filtered = append(filtered, diff)
		}
	}
	return filtered
}

// CompareConfigs compares two nginx configurations and returns the differences
func CompareConfigs(oldConfig, newConfig *config.Config) *DiffResult {
	differ := &configDiffer{
		result: &DiffResult{
			Differences: make([]Difference, 0),
			Summary:     DiffSummary{},
		},
	}

	differ.compareBlocks("", oldConfig.Block, newConfig.Block)
	differ.calculateSummary()

	return differ.result
}

// CompareConfigStrings compares two configuration strings
func CompareConfigStrings(oldConfigStr, newConfigStr string) (*DiffResult, error) {
	// Parse both configurations
	oldParser := &simpleParser{content: oldConfigStr}
	newParser := &simpleParser{content: newConfigStr}

	oldConfig, err := oldParser.parseToMap()
	if err != nil {
		return nil, fmt.Errorf("failed to parse old config: %w", err)
	}

	newConfig, err := newParser.parseToMap()
	if err != nil {
		return nil, fmt.Errorf("failed to parse new config: %w", err)
	}

	return CompareConfigMaps(oldConfig, newConfig), nil
}

// CompareConfigMaps compares two configuration maps
func CompareConfigMaps(oldConfig, newConfig map[string]interface{}) *DiffResult {
	differ := &mapDiffer{
		result: &DiffResult{
			Differences: make([]Difference, 0),
			Summary:     DiffSummary{},
		},
	}

	differ.compareMaps("", oldConfig, newConfig)
	differ.calculateSummary()

	return differ.result
}

// configDiffer handles comparison between config.Config objects
type configDiffer struct {
	result *DiffResult
}

func (cd *configDiffer) compareBlocks(path string, oldBlock, newBlock config.IBlock) {
	if oldBlock == nil && newBlock == nil {
		return
	}

	if oldBlock == nil {
		cd.addDifference(DiffAdded, path, "block", "", "added", "")
		return
	}

	if newBlock == nil {
		cd.addDifference(DiffRemoved, path, "block", "removed", "", "")
		return
	}

	// Compare directives
	oldDirectives := cd.getDirectiveMap(oldBlock.GetDirectives())
	newDirectives := cd.getDirectiveMap(newBlock.GetDirectives())

	// Find added and modified directives
	for name, newDir := range newDirectives {
		if oldDir, exists := oldDirectives[name]; exists {
			cd.compareDirectives(path, oldDir, newDir)
		} else {
			cd.addDirectiveDifference(DiffAdded, path, newDir, "", cd.getDirectiveValue(newDir))
		}
	}

	// Find removed directives
	for name, oldDir := range oldDirectives {
		if _, exists := newDirectives[name]; !exists {
			cd.addDirectiveDifference(DiffRemoved, path, oldDir, cd.getDirectiveValue(oldDir), "")
		}
	}
}

func (cd *configDiffer) compareDirectives(path string, oldDir, newDir config.IDirective) {
	oldValue := cd.getDirectiveValue(oldDir)
	newValue := cd.getDirectiveValue(newDir)

	if oldValue != newValue {
		cd.addDirectiveDifference(DiffModified, path, newDir, oldValue, newValue)
	}

	// Compare blocks if both directives have them
	if oldDir.GetBlock() != nil && newDir.GetBlock() != nil {
		blockPath := cd.buildPath(path, newDir.GetName())
		cd.compareBlocks(blockPath, oldDir.GetBlock(), newDir.GetBlock())
	} else if oldDir.GetBlock() != nil {
		blockPath := cd.buildPath(path, newDir.GetName())
		cd.addDifference(DiffRemoved, blockPath, "block", "removed", "", "")
	} else if newDir.GetBlock() != nil {
		blockPath := cd.buildPath(path, newDir.GetName())
		cd.addDifference(DiffAdded, blockPath, "block", "", "added", "")
	}
}

func (cd *configDiffer) getDirectiveMap(directives []config.IDirective) map[string]config.IDirective {
	directiveMap := make(map[string]config.IDirective)
	for _, dir := range directives {
		// For simplicity, use directive name as key
		// In a more sophisticated implementation, we might include parameters
		key := dir.GetName()
		if len(dir.GetParameters()) > 0 {
			key += ":" + cd.getDirectiveValue(dir)
		}
		directiveMap[key] = dir
	}
	return directiveMap
}

func (cd *configDiffer) getDirectiveValue(dir config.IDirective) string {
	if dir == nil {
		return ""
	}

	params := dir.GetParameters()
	if len(params) == 0 {
		return ""
	}

	values := make([]string, len(params))
	for i, param := range params {
		values[i] = param.GetValue()
	}

	return strings.Join(values, " ")
}

func (cd *configDiffer) addDirectiveDifference(diffType DiffType, path string, dir config.IDirective, oldValue, newValue string) {
	cd.addDifference(diffType, path, dir.GetName(), oldValue, newValue, "")
}

func (cd *configDiffer) addDifference(diffType DiffType, path, directiveName, oldValue, newValue, description string) {
	diff := Difference{
		Type:          diffType,
		Path:          path,
		DirectiveName: directiveName,
		OldValue:      oldValue,
		NewValue:      newValue,
		Description:   description,
	}

	cd.result.Differences = append(cd.result.Differences, diff)
}

func (cd *configDiffer) buildPath(parentPath, name string) string {
	if parentPath == "" {
		return name
	}
	return parentPath + "/" + name
}

func (cd *configDiffer) calculateSummary() {
	summary := DiffSummary{}
	for _, diff := range cd.result.Differences {
		summary.Total++
		switch diff.Type {
		case DiffAdded:
			summary.Added++
		case DiffRemoved:
			summary.Removed++
		case DiffModified:
			summary.Modified++
		case DiffMoved:
			summary.Moved++
		}
	}
	cd.result.Summary = summary
}

// mapDiffer handles comparison between map representations
type mapDiffer struct {
	result *DiffResult
}

func (md *mapDiffer) compareMaps(path string, oldMap, newMap map[string]interface{}) {
	// Get all keys
	allKeys := make(map[string]bool)
	for key := range oldMap {
		allKeys[key] = true
	}
	for key := range newMap {
		allKeys[key] = true
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(allKeys))
	for key := range allKeys {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		oldValue, oldExists := oldMap[key]
		newValue, newExists := newMap[key]

		keyPath := md.buildPath(path, key)

		if !oldExists && newExists {
			md.addDifference(DiffAdded, keyPath, key, "", md.valueToString(newValue), "")
		} else if oldExists && !newExists {
			md.addDifference(DiffRemoved, keyPath, key, md.valueToString(oldValue), "", "")
		} else if oldExists && newExists {
			if md.valueToString(oldValue) != md.valueToString(newValue) {
				md.addDifference(DiffModified, keyPath, key, md.valueToString(oldValue), md.valueToString(newValue), "")
			}

			// Recursively compare nested maps
			if oldSubMap, oldOk := oldValue.(map[string]interface{}); oldOk {
				if newSubMap, newOk := newValue.(map[string]interface{}); newOk {
					md.compareMaps(keyPath, oldSubMap, newSubMap)
				}
			}
		}
	}
}

func (md *mapDiffer) valueToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []string:
		return strings.Join(v, " ")
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (md *mapDiffer) buildPath(parentPath, key string) string {
	if parentPath == "" {
		return key
	}
	return parentPath + "/" + key
}

func (md *mapDiffer) addDifference(diffType DiffType, path, directiveName, oldValue, newValue, description string) {
	diff := Difference{
		Type:          diffType,
		Path:          path,
		DirectiveName: directiveName,
		OldValue:      oldValue,
		NewValue:      newValue,
		Description:   description,
	}

	md.result.Differences = append(md.result.Differences, diff)
}

func (md *mapDiffer) calculateSummary() {
	summary := DiffSummary{}
	for _, diff := range md.result.Differences {
		summary.Total++
		switch diff.Type {
		case DiffAdded:
			summary.Added++
		case DiffRemoved:
			summary.Removed++
		case DiffModified:
			summary.Modified++
		case DiffMoved:
			summary.Moved++
		}
	}
	md.result.Summary = summary
}

// simpleParser is a basic parser for string-based comparison
type simpleParser struct {
	content string
}

func (sp *simpleParser) parseToMap() (map[string]interface{}, error) {
	// This is a simplified parser for demonstration
	// In a real implementation, you would use the actual nginx parser
	result := make(map[string]interface{})

	lines := strings.Split(sp.content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Simple directive parsing
		if strings.Contains(line, " ") {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSuffix(strings.TrimSpace(parts[1]), ";")
				result[key] = value
			}
		}
	}

	return result, nil
}
