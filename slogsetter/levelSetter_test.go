package slogsetter_test

import (
	"log/slog"
	"maps"
	"slices"
	"testing"

	"github.com/nickwells/param.mod/v6/psetter"
	"github.com/nickwells/slogsetter.mod/slogsetter"
	"github.com/nickwells/testhelper.mod/v2/testhelper"
)

func TestLevelPopulateDefaultLevelMap(t *testing.T) {
	var nilMap slogsetter.LevelMap

	testCases := []struct {
		testhelper.ID
		testhelper.ExpPanic
		m        *slogsetter.LevelMap
		preFunc  func(*slogsetter.LevelMap)
		postFunc func(*slogsetter.LevelMap)
		expMap   slogsetter.LevelMap
	}{
		{
			ID: testhelper.MkID("just the default"),
			m:  &slogsetter.LevelMap{},
			expMap: slogsetter.LevelMap{
				slog.LevelDebug.String(): slog.LevelDebug,
				slog.LevelInfo.String():  slog.LevelInfo,
				slog.LevelWarn.String():  slog.LevelWarn,
				slog.LevelError.String(): slog.LevelError,
			},
		},
		{
			ID: testhelper.MkID("nil map - gets created"),
			m:  &nilMap,
			expMap: slogsetter.LevelMap{
				slog.LevelDebug.String(): slog.LevelDebug,
				slog.LevelInfo.String():  slog.LevelInfo,
				slog.LevelWarn.String():  slog.LevelWarn,
				slog.LevelError.String(): slog.LevelError,
			},
		},
		{
			ID: testhelper.MkID("nil pointer - panics"),
			ExpPanic: testhelper.MkExpPanic(
				"You must call PopulateDefaultLevelMap" +
					" with a non-nil pointer"),
		},
		{
			ID: testhelper.MkID("set an entry before - gets overwritten"),
			m:  &slogsetter.LevelMap{},
			preFunc: func(m *slogsetter.LevelMap) {
				(*m)[slog.LevelDebug.String()] = slog.Level(42)
			},
			expMap: slogsetter.LevelMap{
				slog.LevelDebug.String(): slog.LevelDebug,
				slog.LevelInfo.String():  slog.LevelInfo,
				slog.LevelWarn.String():  slog.LevelWarn,
				slog.LevelError.String(): slog.LevelError,
			},
		},
		{
			ID: testhelper.MkID("set an entry after - doesn't get overwritten"),
			m:  &slogsetter.LevelMap{},
			postFunc: func(m *slogsetter.LevelMap) {
				(*m)[slog.LevelDebug.String()] = slog.Level(42)
			},
			expMap: slogsetter.LevelMap{
				slog.LevelDebug.String(): slog.Level(42),
				slog.LevelInfo.String():  slog.LevelInfo,
				slog.LevelWarn.String():  slog.LevelWarn,
				slog.LevelError.String(): slog.LevelError,
			},
		},
		{
			ID: testhelper.MkID("new entry before - doesn't get overwritten"),
			m:  &slogsetter.LevelMap{},
			preFunc: func(m *slogsetter.LevelMap) {
				(*m)["Forty-two"] = slog.Level(42)
			},
			expMap: slogsetter.LevelMap{
				"Forty-two":              slog.Level(42),
				slog.LevelDebug.String(): slog.LevelDebug,
				slog.LevelInfo.String():  slog.LevelInfo,
				slog.LevelWarn.String():  slog.LevelWarn,
				slog.LevelError.String(): slog.LevelError,
			},
		},
		{
			ID: testhelper.MkID("new entry after - doesn't get overwritten"),
			m:  &slogsetter.LevelMap{},
			postFunc: func(m *slogsetter.LevelMap) {
				(*m)["Forty-two"] = slog.Level(42)
			},
			expMap: slogsetter.LevelMap{
				"Forty-two":              slog.Level(42),
				slog.LevelDebug.String(): slog.LevelDebug,
				slog.LevelInfo.String():  slog.LevelInfo,
				slog.LevelWarn.String():  slog.LevelWarn,
				slog.LevelError.String(): slog.LevelError,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.preFunc != nil {
				tc.preFunc(tc.m)
			}

			panicked, panicVal := testhelper.PanicSafe(
				func() {
					slogsetter.PopulateDefaultLevelMap(tc.m)
				})

			testhelper.CheckExpPanic(t, panicked, panicVal, tc)

			if tc.postFunc != nil {
				tc.postFunc(tc.m)
			}

			if tc.m != nil {
				if !maps.Equal(*tc.m, tc.expMap) {
					t.Log(tc.IDStr())
					t.Log("\t: tc.m: ", *tc.m)
					t.Log("\t: expected: ", tc.expMap)
					t.Errorf("\t: maps differ\n")
				}
			}
		})
	}
}

func TestLevelSetWithVal(t *testing.T) {
	testCases := []struct {
		testhelper.ID
		testhelper.ExpErr
		val      string
		setter   slogsetter.Level
		expLevel slog.Level
	}{
		{
			ID:       testhelper.MkID("good level name"),
			val:      "INFO",
			setter:   slogsetter.Level{},
			expLevel: slog.LevelInfo,
		},
		{
			ID:       testhelper.MkID("bad level name"),
			ExpErr:   testhelper.MkExpErr(`bad logging level name ("blah")`),
			val:      "blah",
			setter:   slogsetter.Level{},
			expLevel: slog.LevelInfo,
		},
		{
			ID: testhelper.MkID("bad level name with alternative"),
			ExpErr: testhelper.MkExpErr(
				`bad logging level name ("ERRORx")`,
				`, did you mean "ERROR"?`),
			val:      "ERRORx",
			setter:   slogsetter.Level{},
			expLevel: slog.LevelInfo,
		},
		{
			ID:  testhelper.MkID("good level name with alternative LevelMap"),
			val: "Forty-two",
			setter: slogsetter.Level{
				LevelMap: slogsetter.LevelMap{
					"Forty-two":   slog.Level(42),
					"Forty-three": slog.Level(43),
				},
			},
			expLevel: slog.Level(42),
		},
		{
			ID:     testhelper.MkID("bad level name with alternative LevelMap"),
			ExpErr: testhelper.MkExpErr(`bad logging level name ("ERRORx")`),
			val:    "ERRORx",
			setter: slogsetter.Level{
				LevelMap: slogsetter.LevelMap{
					"Forty-two":   slog.Level(42),
					"Forty-three": slog.Level(43),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			lvl := slog.LevelInfo
			tc.setter.Value = &lvl
			err := tc.setter.SetWithVal("", tc.val)

			testhelper.CheckExpErr(t, err, tc)

			if lvl != tc.expLevel {
				t.Log(tc.IDStr())
				t.Log("\t:   actual level: ", lvl)
				t.Log("\t: expected level: ", tc.expLevel)
				t.Errorf("\t: level has not been set correctly\n")
			}
		})
	}
}

func TestLevelAllowedValues(t *testing.T) {
	setter := slogsetter.Level{}
	if av := setter.AllowedValues(); av != "A logging level name. " {
		t.Errorf("AllowedValues returned an unexpected value: %q", av)
	}
}

func TestLevelValDescribe(t *testing.T) {
	setter := slogsetter.Level{}
	if av := setter.ValDescribe(); av != "log-level" {
		t.Errorf("ValDescribe returned an unexpected value: %q", av)
	}
}

func TestLevelAllowedValuesMap(t *testing.T) {
	var dfltLevelsOnly,
		dfltLevelsWithDups slogsetter.LevelMap

	slogsetter.PopulateDefaultLevelMap(&dfltLevelsOnly)
	slogsetter.PopulateDefaultLevelMap(&dfltLevelsWithDups)
	dfltLevelsWithDups["a"] = slog.LevelInfo
	dfltLevelsWithDups["b"] = slog.LevelInfo
	dfltLevelsWithDups["c"] = slog.LevelError
	dfltLevelsWithDups["d"] = slog.LevelWarn
	dfltLevelsWithDups["aPlusOne"] = slog.LevelInfo + 1

	testCases := []struct {
		testhelper.ID
		setter          slogsetter.Level
		expAValMap      psetter.AllowedVals[string]
		expAValAliasMap psetter.Aliases[string]
	}{
		{
			ID: testhelper.MkID("no defaults, just one entry"),
			setter: slogsetter.Level{
				LevelMap: slogsetter.LevelMap{
					"Forty-two": slog.Level(42),
				},
			},
			expAValMap: psetter.AllowedVals[string]{
				"Forty-two": "ERROR+34",
			},
		},
		{
			ID: testhelper.MkID("no defaults, two entries"),
			setter: slogsetter.Level{
				LevelMap: slogsetter.LevelMap{
					"Forty-two":   slog.Level(42),
					"Forty-three": slog.Level(43),
				},
			},
			expAValMap: psetter.AllowedVals[string]{
				"Forty-two":   "ERROR+34",
				"Forty-three": "ERROR+35",
			},
		},
		{
			ID: testhelper.MkID("no defaults, 3 entries (only 2 levels)"),
			setter: slogsetter.Level{
				LevelMap: slogsetter.LevelMap{
					"a": slog.Level(42),
					"b": slog.Level(43),
					"x": slog.Level(43),
				},
			},
			expAValMap: psetter.AllowedVals[string]{
				"a": "ERROR+34",
				"b": "ERROR+35",
			},
			expAValAliasMap: psetter.Aliases[string]{
				"x": []string{"b"},
			},
		},
		{
			ID: testhelper.MkID("only defaults"),
			setter: slogsetter.Level{
				LevelMap: dfltLevelsOnly,
			},
			expAValMap: psetter.AllowedVals[string]{
				slog.LevelDebug.String(): slog.LevelDebug.String(),
				slog.LevelInfo.String():  slog.LevelInfo.String(),
				slog.LevelWarn.String():  slog.LevelWarn.String(),
				slog.LevelError.String(): slog.LevelError.String(),
			},
		},
		{
			ID: testhelper.MkID("defaults with aliases and extras"),
			setter: slogsetter.Level{
				LevelMap: dfltLevelsWithDups,
			},
			expAValMap: psetter.AllowedVals[string]{
				slog.LevelDebug.String(): slog.LevelDebug.String(),
				slog.LevelInfo.String():  slog.LevelInfo.String(),
				slog.LevelWarn.String():  slog.LevelWarn.String(),
				slog.LevelError.String(): slog.LevelError.String(),
				"aPlusOne":               (slog.LevelInfo + 1).String(),
			},
			expAValAliasMap: psetter.Aliases[string]{
				"a": []string{slog.LevelInfo.String()},
				"b": []string{slog.LevelInfo.String()},
				"c": []string{slog.LevelError.String()},
				"d": []string{slog.LevelWarn.String()},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			avm := tc.setter.AllowedValuesMap()
			if !maps.Equal(avm, tc.expAValMap) {
				t.Log(tc.IDStr())
				t.Log("\t: expected allowed vals map:\t", tc.expAValMap)
				t.Log("\t:   actual allowed vals map:\t", avm)
				t.Error("\t: the allowed values maps differ")
			}

			avam := tc.setter.AllowedValuesAliasMap()
			if !maps.EqualFunc(avam, tc.expAValAliasMap, slices.Equal) {
				t.Log(tc.IDStr())
				t.Log("\t: expected allowed vals alias map:\t",
					tc.expAValAliasMap)
				t.Log("\t:   actual allowed vals alias map:\t", avam)
				t.Error("\t: the allowed values maps differ")
			}
		})
	}
}

func TestLevelCheckSetter(t *testing.T) {
	var v slog.Level

	emptyLevelMap := slogsetter.LevelMap{}
	oneEntryLevelMap := slogsetter.LevelMap{
		"a": slog.LevelInfo,
	}

	testCases := []struct {
		testhelper.ID
		testhelper.ExpPanic
		setter slogsetter.Level
	}{
		{
			ID: testhelper.MkID("nothing set - first check: Value"),
			ExpPanic: testhelper.MkExpPanic(
				"setterName: slogsetter.Level Check failed",
				"the Value to be set is nil",
			),
			setter: slogsetter.Level{},
		},
		{
			ID: testhelper.MkID("Value set - empty LevelMap"),
			ExpPanic: testhelper.MkExpPanic(
				"setterName: slogsetter.Level Check failed",
				"a LevelMap has been set but it is empty",
			),
			setter: slogsetter.Level{
				Value:    &v,
				LevelMap: emptyLevelMap,
			},
		},
		{
			ID: testhelper.MkID("Value set - one entry LevelMap"),
			ExpPanic: testhelper.MkExpPanic(
				"setterName: slogsetter.Level Check failed",
				"a LevelMap has been set but it has only one entry",
			),
			setter: slogsetter.Level{
				Value:    &v,
				LevelMap: oneEntryLevelMap,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			panicked, panickVal := testhelper.PanicSafe(func() {
				tc.setter.CheckSetter("setterName")
			})
			testhelper.CheckExpPanic(t, panicked, panickVal, tc)
		})
	}
}

func TestLevelCurrentValue(t *testing.T) {
	testCases := []struct {
		testhelper.ID
		v            slog.Level
		expStringVal string
	}{
		{
			ID:           testhelper.MkID("CurrentValue: INFO"),
			v:            slog.LevelInfo,
			expStringVal: "INFO",
		},
		{
			ID:           testhelper.MkID("CurrentValue: INFO+1"),
			v:            slog.LevelInfo + 1,
			expStringVal: "INFO+1",
		},
		{
			ID:           testhelper.MkID("CurrentValue: DEBUG-99"),
			v:            slog.LevelDebug - 99,
			expStringVal: "DEBUG-99",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			s := slogsetter.Level{
				Value: &tc.v,
			}
			strVal := s.CurrentValue()
			testhelper.DiffString(t,
				tc.IDStr(), "value",
				strVal, tc.expStringVal)
		})
	}
}
