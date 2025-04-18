package testtools

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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

func Load[T any](filename string) T {
	var target T

	dir := "testdata/" + previousFuncName(2)
	inPath := "/input/" + dir + filename + ".json"

	file, err := os.Open(inPath)
	CrashOn(err, "failed to open "+inPath)
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&target)
	CrashOn(err, "failed to decode data")
	return target
}

func Init() {
	inputDir := "testdata/" + previousFuncName(2) + "/input"
	err := os.MkdirAll(inputDir, os.ModePerm)
	CrashOn(err, "failed to make directory")

	outputDir := "testdata/" + previousFuncName(2) + "/output"
	err = os.MkdirAll(outputDir, os.ModePerm)
	CrashOn(err, "failed to make output directory")
}

func Save(data any, filename string) error {
	outputDir := "testdata/" + previousFuncName(2) + "/output"

	err := os.MkdirAll(outputDir, os.ModePerm)
	CrashOn(err, "failed to make directory")

	outPath := outputDir + "/" + filename + ".json"

	file, err := os.Create(outPath)
	CrashOn(err, "failed to create "+outPath)
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(data)
}

func previousFuncName(backSteps int) string {
	pc, _, _, _ := runtime.Caller(backSteps)
	full := runtime.FuncForPC(pc).Name() // e.g., "main.myFunction"
	parts := strings.Split(full, ".")
	name := parts[len(parts)-1] // "myFunction"
	return name
}

// ================================================
// RECORD OR ASSERT
// ================================================

// RecordOrAssert saves test data if it doesn't exist, or compares it otherwise.
// It uses the calling test function's name as the filename if none is given.
func RecordOrAssert[T any](t *testing.T, filename string, got T) {
	t.Helper()
	outputDir := "testdata/" + previousFuncName(2) + "/output"
	CrashOn(os.MkdirAll(outputDir, 0755), "failed to make testData dir")

	var expected T
	err := fromFile(outputDir, filename, &expected)
	if err != nil {
		CrashOn(toFile(outputDir, filename, got), "failed to load file")
		return
	}

	normalize := func(v any) any {
		encoded, err := json.Marshal(v)
		CrashOn(err, "failed to marshal data")

		var decoded any
		dec := json.NewDecoder(bytes.NewReader(encoded))
		dec.UseNumber()
		CrashOn(dec.Decode(&decoded), "failed to decode data")
		return normalizeJSON(decoded)
	}

	gotNorm := normalize(got)
	expectedNorm := normalize(expected)

	actualFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + "_actual"
	actualPath := filepath.Join(outputDir, actualFilename)

	if !reflect.DeepEqual(gotNorm, expectedNorm) {
		diff := cmp.Diff(expectedNorm, gotNorm)
		t.Errorf("Mismatch with testdata %s:\nDiff:\n%s", filename, diff)
		CrashOn(toFile(outputDir, actualFilename, got), "failed to save data")
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
