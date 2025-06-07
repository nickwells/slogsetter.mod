package slogsetter

import (
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"sort"

	"github.com/nickwells/param.mod/v6/psetter"
)

// LevelMap is the type of a map between names and slog.Level
type LevelMap map[string]slog.Level

var dfltLevelMap = LevelMap{
	slog.LevelDebug.String(): slog.LevelDebug,
	slog.LevelInfo.String():  slog.LevelInfo,
	slog.LevelWarn.String():  slog.LevelWarn,
	slog.LevelError.String(): slog.LevelError,
}

// Level is a parameter setter used to populate slog.Level values.
type Level struct {
	psetter.ValueReqMandatory

	// Value is a pointer to the slog Level value that this setter will set
	Value *slog.Level
	// LevelMap is an alternative map of names to logging levels. It if
	// is not set then the default map will be used.
	LevelMap LevelMap
}

// PopulateDefaultLevelMap initialises the map, if necessary, and then copies
// in the values from the default map. This can be used to add extra levels
// to the default map or to add aliases to the entries in the default map.
//
// Note that any entries in the map supplied with the same key as the value
// in the default map will be overwritten with the default values. To
// supersede the default entries set the new values after calling this
// function.
func PopulateDefaultLevelMap(m *LevelMap) {
	if m == nil {
		panic("You must call PopulateDefaultLevelMap with a non-nil pointer")
	}

	if *m == nil {
		*m = make(LevelMap)
	}

	maps.Copy(*m, dfltLevelMap)
}

// effectiveLevelMap returns the LevelMap that the setter will use
func (s Level) effectiveLevelMap() LevelMap {
	levelMap := dfltLevelMap

	if s.LevelMap != nil {
		levelMap = s.LevelMap
	}

	return levelMap
}

// SetWithVal (called with the value following the parameter) makes sure that
// the supplied paramVal is a valid entry in the LevelMap and if so it sets
// the Value. Otherwise it returns an error.
func (s Level) SetWithVal(_ string, paramVal string) error {
	levelMap := s.effectiveLevelMap()

	level, ok := levelMap[paramVal]
	if !ok {
		return fmt.Errorf("bad logging level name (%q)%s",
			paramVal,
			psetter.SuggestionString(
				psetter.SuggestedVals(
					paramVal,
					slices.Collect(maps.Keys(levelMap)),
				),
			),
		)
	}

	*s.Value = level

	return nil
}

// AllowedValues returns a string describing the allowed values
func (s Level) AllowedValues() string {
	return "A logging level name. "
}

// sortedLevelNames returns the names of the levels in the levelMap sorted by:
//   - the numeric level number
//   - whether the name is in the default map
//   - alphabetically
func (s Level) sortedLevelNames() []string {
	levelMap := s.effectiveLevelMap()

	names := slices.Collect(maps.Keys(levelMap))
	if len(names) == 0 {
		return names
	}

	sort.Slice(names, func(i, j int) bool {
		// sort by level number ...
		if levelMap[names[i]] < levelMap[names[j]] {
			return true
		}

		if levelMap[names[j]] < levelMap[names[i]] {
			return false
		}

		// ... then entries in the default map ...
		if _, ok := dfltLevelMap[names[i]]; ok {
			return true
		}

		if _, ok := dfltLevelMap[names[j]]; ok {
			return false
		}

		// ... and finally alphabetically
		return names[i] < names[j]
	})

	return names
}

// AllowedValuesMap returns a map of logging level names to the corresponding
// level value
func (s Level) AllowedValuesMap() psetter.AllowedVals[string] {
	levelMap := s.effectiveLevelMap()

	names := s.sortedLevelNames()
	if len(names) == 0 {
		return psetter.AllowedVals[string]{}
	}

	avm := make(psetter.AllowedVals[string])
	lastLevel := levelMap[names[0]] - 1

	for _, name := range names {
		if levelMap[name] != lastLevel {
			avm[name] = levelMap[name].String()
		}

		lastLevel = levelMap[name]
	}

	return avm
}

// AllowedValuesAliasMap returns a map of logging level aliases to the
// logging level names
func (s Level) AllowedValuesAliasMap() psetter.Aliases[string] {
	levelMap := s.effectiveLevelMap()

	names := s.sortedLevelNames()
	if len(names) == 0 {
		return psetter.Aliases[string]{}
	}

	avam := make(psetter.Aliases[string])
	lastLevel := levelMap[names[0]] - 1
	aliasTo := ""

	for _, name := range names {
		if levelMap[name] != lastLevel {
			aliasTo = name
		} else {
			avam[name] = []string{aliasTo}
		}

		lastLevel = levelMap[name]
	}

	return avam
}

// ValDescribe returns a string describing the value that can follow the
// parameter
func (s Level) ValDescribe() string {
	return "log-level"
}

// CurrentValue returns the current setting of the parameter value
func (s Level) CurrentValue() string {
	return s.Value.String()
}

// CheckSetter panics if the setter has not been properly created - if the
// Value is nil or if the level map has been set and is empty or has only a
// single entry.
func (s Level) CheckSetter(name string) {
	intro := name + ": slogsetter.Level Check failed:"

	if s.Value == nil {
		panic(intro + " the Value to be set is nil")
	}

	levelMap := s.effectiveLevelMap()

	if len(levelMap) == 0 {
		panic(intro + " a LevelMap has been set but it is empty")
	}

	if len(levelMap) == 1 {
		panic(intro + " a LevelMap has been set but it has only one entry")
	}
}
