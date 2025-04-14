package testtools

import (
	"encoding/json"
	"os"
	"testing"

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
