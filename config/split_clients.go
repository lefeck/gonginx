package config

import (
	"errors"
	"strconv"
	"strings"
)

// SplitClients represents a split_clients block in nginx configuration
// split_clients $variable $mapped_variable { ... }
type SplitClients struct {
	Variable       string // Source variable (e.g., $remote_addr)
	MappedVariable string // Target variable (e.g., $variant)
	Entries        []*SplitClientsEntry
	Comment        []string
	DefaultInlineComment
	Parent IDirective
	Line   int
}

// SplitClientsEntry represents a single entry in a split_clients block
type SplitClientsEntry struct {
	Percentage string // Percentage (e.g., "0.5%", "2.0%", "*")
	Value      string // Value to set for this percentage
	Comment    []string
	DefaultInlineComment
	Parent IDirective
	Line   int
}

// SetLine sets the line number
func (sc *SplitClients) SetLine(line int) {
	sc.Line = line
}

// GetLine returns the line number
func (sc *SplitClients) GetLine() int {
	return sc.Line
}

// SetParent sets the parent directive
func (sc *SplitClients) SetParent(parent IDirective) {
	sc.Parent = parent
}

// GetParent returns the parent directive
func (sc *SplitClients) GetParent() IDirective {
	return sc.Parent
}

// SetComment sets the directive comment
func (sc *SplitClients) SetComment(comment []string) {
	sc.Comment = comment
}

// GetName implements the IDirective interface
func (sc *SplitClients) GetName() string {
	return "split_clients"
}

// GetParameters returns the split_clients parameters (source and target variables)
func (sc *SplitClients) GetParameters() []Parameter {
	return []Parameter{
		{Value: sc.Variable},
		{Value: sc.MappedVariable},
	}
}

// GetBlock returns the split_clients itself, which implements IBlock
func (sc *SplitClients) GetBlock() IBlock {
	return sc
}

// GetComment returns the directive comment
func (sc *SplitClients) GetComment() []string {
	return sc.Comment
}

// GetDirectives returns the split_clients entries as directives
func (sc *SplitClients) GetDirectives() []IDirective {
	directives := make([]IDirective, len(sc.Entries))
	for i, entry := range sc.Entries {
		directives[i] = entry
	}
	return directives
}

// GetCodeBlock returns empty string (not a literal code block)
func (sc *SplitClients) GetCodeBlock() string {
	return ""
}

// FindDirectives finds directives in the split_clients block
func (sc *SplitClients) FindDirectives(directiveName string) []IDirective {
	// Split_clients blocks typically only contain percentage entries
	// but we implement this for consistency
	var directives []IDirective
	for _, entry := range sc.Entries {
		if entry.GetName() == directiveName {
			directives = append(directives, entry)
		}
	}
	return directives
}

// AddEntry adds a new percentage split to the split_clients block
func (sc *SplitClients) AddEntry(percentage, value string) error {
	// Validate percentage format
	if err := sc.validatePercentage(percentage); err != nil {
		return err
	}

	entry := &SplitClientsEntry{
		Percentage: percentage,
		Value:      value,
		Parent:     sc,
	}
	sc.Entries = append(sc.Entries, entry)
	return nil
}

// validatePercentage checks if the percentage format is valid
func (sc *SplitClients) validatePercentage(percentage string) error {
	// Special case for wildcard
	if percentage == "*" {
		return nil
	}

	// Must end with %
	if !strings.HasSuffix(percentage, "%") {
		return errors.New("percentage must end with %")
	}

	// Remove % and try to parse as float
	percentStr := strings.TrimSuffix(percentage, "%")
	percent, err := strconv.ParseFloat(percentStr, 64)
	if err != nil {
		return errors.New("invalid percentage format: " + percentage)
	}

	// Check reasonable range (0 to 100)
	if percent < 0 || percent > 100 {
		return errors.New("percentage must be between 0% and 100%")
	}

	return nil
}

// GetTotalPercentage calculates the total percentage allocated
func (sc *SplitClients) GetTotalPercentage() (float64, error) {
	total := 0.0

	for _, entry := range sc.Entries {
		if entry.Percentage == "*" {
			continue // Wildcard doesn't count towards total
		}

		percentStr := strings.TrimSuffix(entry.Percentage, "%")
		percent, err := strconv.ParseFloat(percentStr, 64)
		if err != nil {
			return 0, err
		}
		total += percent
	}

	return total, nil
}

// HasWildcard checks if there's a wildcard entry (*)
func (sc *SplitClients) HasWildcard() bool {
	for _, entry := range sc.Entries {
		if entry.Percentage == "*" {
			return true
		}
	}
	return false
}

// GetWildcardValue returns the value for the wildcard entry
func (sc *SplitClients) GetWildcardValue() string {
	for _, entry := range sc.Entries {
		if entry.Percentage == "*" {
			return entry.Value
		}
	}
	return ""
}

// RemoveEntry removes an entry by percentage
func (sc *SplitClients) RemoveEntry(percentage string) bool {
	for i, entry := range sc.Entries {
		if entry.Percentage == percentage {
			sc.Entries = append(sc.Entries[:i], sc.Entries[i+1:]...)
			return true
		}
	}
	return false
}

// GetEntriesByValue returns all entries with a specific value
func (sc *SplitClients) GetEntriesByValue(value string) []*SplitClientsEntry {
	var entries []*SplitClientsEntry
	for _, entry := range sc.Entries {
		if entry.Value == value {
			entries = append(entries, entry)
		}
	}
	return entries
}

// NewSplitClients creates a new SplitClients from a directive
func NewSplitClients(directive IDirective) (*SplitClients, error) {
	parameters := directive.GetParameters()
	if len(parameters) < 2 {
		return nil, errors.New("split_clients directive requires exactly 2 parameters: source and target variables")
	}

	if len(parameters) > 2 {
		return nil, errors.New("split_clients directive accepts exactly 2 parameters only")
	}

	splitClientsBlock := &SplitClients{
		Variable:             parameters[0].GetValue(),
		MappedVariable:       parameters[1].GetValue(),
		Entries:              make([]*SplitClientsEntry, 0),
		Comment:              directive.GetComment(),
		DefaultInlineComment: DefaultInlineComment{InlineComment: directive.GetInlineComment()},
	}

	if directive.GetBlock() == nil {
		return nil, errors.New("split_clients directive must have a block")
	}

	// Parse split_clients entries from the block
	for _, d := range directive.GetBlock().GetDirectives() {
		entry, err := NewSplitClientsEntry(d)
		if err != nil {
			return nil, err
		}
		entry.SetParent(splitClientsBlock)
		entry.SetLine(d.GetLine())
		splitClientsBlock.Entries = append(splitClientsBlock.Entries, entry)
	}

	// Validate the entries
	if err := splitClientsBlock.validateEntries(); err != nil {
		return nil, err
	}

	return splitClientsBlock, nil
}

// validateEntries validates the split_clients entries
func (sc *SplitClients) validateEntries() error {
	if len(sc.Entries) == 0 {
		return nil // Empty block is allowed
	}

	// Check for total percentage > 100%
	total, err := sc.GetTotalPercentage()
	if err != nil {
		return err
	}

	if total > 100.0 {
		return errors.New("total percentage cannot exceed 100%")
	}

	// If there's no wildcard and total < 100%, warn but don't error
	// This is valid nginx configuration (remaining traffic gets empty value)

	return nil
}

// SplitClientsEntry methods implementing IDirective interface

// SetLine sets the line number for SplitClientsEntry
func (sce *SplitClientsEntry) SetLine(line int) {
	sce.Line = line
}

// GetLine returns the line number for SplitClientsEntry
func (sce *SplitClientsEntry) GetLine() int {
	return sce.Line
}

// SetParent sets the parent directive for SplitClientsEntry
func (sce *SplitClientsEntry) SetParent(parent IDirective) {
	sce.Parent = parent
}

// GetParent returns the parent directive for SplitClientsEntry
func (sce *SplitClientsEntry) GetParent() IDirective {
	return sce.Parent
}

// SetComment sets the comment for SplitClientsEntry
func (sce *SplitClientsEntry) SetComment(comment []string) {
	sce.Comment = comment
}

// GetName returns the percentage as the "directive name" for SplitClientsEntry
func (sce *SplitClientsEntry) GetName() string {
	return sce.Percentage
}

// GetParameters returns the value as parameter for SplitClientsEntry
func (sce *SplitClientsEntry) GetParameters() []Parameter {
	return []Parameter{{Value: sce.Value}}
}

// GetBlock returns nil as SplitClientsEntry doesn't have a block
func (sce *SplitClientsEntry) GetBlock() IBlock {
	return nil
}

// GetComment returns the comment for SplitClientsEntry
func (sce *SplitClientsEntry) GetComment() []string {
	return sce.Comment
}

// NewSplitClientsEntry creates a new SplitClientsEntry from a directive
func NewSplitClientsEntry(directive IDirective) (*SplitClientsEntry, error) {
	parameters := directive.GetParameters()
	if len(parameters) < 1 {
		return nil, errors.New("split_clients entry must have at least one parameter (value)")
	}

	// The percentage is the directive name, value is the first parameter
	entry := &SplitClientsEntry{
		Percentage:           directive.GetName(),
		Value:                parameters[0].GetValue(),
		Comment:              directive.GetComment(),
		DefaultInlineComment: DefaultInlineComment{InlineComment: directive.GetInlineComment()},
	}

	return entry, nil
}
