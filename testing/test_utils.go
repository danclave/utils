package testtools

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog/log"
)

func SkipIff(t *testing.T, condition bool) {
	if condition {
		t.SkipNow()
	}
}

func RunIf(t *testing.T, condition bool) {
	if !condition {
		t.SkipNow()
	}
}

type TestDataStore struct {
	BaseDir string
}

// ================================================
// TEST DATA
// ================================================

func Load[T any](dir string, filename string) T {
	var target T

	inPath := "testData/" + dir + "/input/" + filename + ".json"

	file, err := os.Open(inPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&target)
	return target
}

func Save(dir string, data any, filename string) error {
	outputDir := "testData/" + dir + "/output"

	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Str("dir", outputDir).Msg("Failed to create directories")
		return err
	}

	outPath := outputDir + "/" + filename + ".json"

	file, err := os.Create(outPath)
	if err != nil {
		log.Fatal().Err(err).Str("path", outPath).Msg("Failed to create file")
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(data)
}

// ================================================
// RECORD OR ASSERT
// ================================================

// RecordOrAssert saves test data if it doesn't exist, or compares it otherwise.
// It uses the calling test function's name as the filename if none is given.
func RecordOrAssert[T any](t *testing.T, filename string, got T) {
	t.Helper()

	if filename == "" {
		filename = GetFuncName()
	}

	dir := "testdata"
	CrashOn(os.MkdirAll(dir, 0755))

	var expected T
	err := fromFile(dir, filename, &expected)
	if err != nil {
		CrashOn(toFile(dir, filename, got))
		return
	}

	normalize := func(v any) any {
		encoded, err := json.Marshal(v)
		CrashOn(err)

		var decoded any
		dec := json.NewDecoder(bytes.NewReader(encoded))
		dec.UseNumber()
		CrashOn(dec.Decode(&decoded))
		return normalizeJSON(decoded)
	}

	gotNorm := normalize(got)
	expectedNorm := normalize(expected)

	actualFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + "_actual"
	actualPath := filepath.Join(dir, actualFilename)

	if !reflect.DeepEqual(gotNorm, expectedNorm) {
		diff := cmp.Diff(expectedNorm, gotNorm)
		t.Errorf("Mismatch with testdata %s:\nDiff:\n%s", filename, diff)
		CrashOn(toFile(dir, actualFilename, got))
	} else {
		_ = os.Remove(actualPath + filepath.Ext(filename))
	}
}

// normalizeJSON recursively sorts maps and arrays by their JSON-encoded string representation.
func normalizeJSON(v any) any {
	switch val := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make(map[string]any, len(val))
		for _, k := range keys {
			out[k] = normalizeJSON(val[k])
		}
		return out
	case []any:
		for i := range val {
			val[i] = normalizeJSON(val[i])
		}
		sort.SliceStable(val, func(i, j int) bool {
			a, _ := json.Marshal(val[i])
			b, _ := json.Marshal(val[j])
			return string(a) < string(b)
		})
		return val
	default:
		return val
	}
}
