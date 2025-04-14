package testtools

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

// I AM UNSAFE BY DEFAULT
func TrustMe[T any](val T, err error) T {
	if err != nil {
		log.Fatal().Err(err).Msg("TrustMe fatal error")
	}
	return val
}

func DoOrDie[T any](val T, err error) T {
	CrashOn(err)
	return val
}

func CrashOn(err error) {
	if err != nil {
		log.Fatal().Err(err).Msg("Crashed due to error")
		panic("Crashed due to error")
	}
}

// DATES
func ToDate(dateString string) time.Time {
	format := "2006-01-02"
	return DoOrDie(time.Parse(format, dateString))
}

func DatesBetween(start, end time.Time) []time.Time {
	var dates []time.Time
	start = start.Truncate(24 * time.Hour)
	end = end.Truncate(24 * time.Hour)

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d)
	}
	return dates
}

func toFile(dir, filename string, data any) error {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Str("dir", dir).Msg("Failed to create directory")
		return err
	}

	outPath := dir + "/" + filename + ".json"

	file, err := os.Create(outPath)
	if err != nil {
		log.Fatal().Err(err).Str("path", outPath).Msg("Failed to create file")
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(data)
}

func fromFile(dir, filename string, target any) error {
	inPath := dir + "/" + filename + ".json"

	file, err := os.Open(inPath)
	if err != nil {
		log.Error().Err(err).Str("path", inPath).Msg("Failed to open file")
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(target)
}

func GetFuncName() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return ""
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return ""
	}
	return fn.Name()
}

func LogLen[T any](label string, items []T) {
	log.Info().Str("label", label).Int("count", len(items)).Msg("Logging slice length")
}

// UnescapeJSON parses either a raw JSON object or a JSON-escaped JSON string into a map
func UnescapeJSON(input string) (map[string]interface{}, error) {
	var result map[string]interface{}

	if err := json.Unmarshal([]byte(input), &result); err == nil {
		return result, nil
	}

	var inner string
	if err := json.Unmarshal([]byte(input), &inner); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input as JSON string: %w", err)
	}

	if err := json.Unmarshal([]byte(inner), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal unescaped JSON: %w", err)
	}

	return result, nil
}
