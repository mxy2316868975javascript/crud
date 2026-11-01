package seed

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/mxy2316868975javascript/crud/internal/domain"
	"github.com/mxy2316868975javascript/crud/internal/repository/postgres"
)

//go:embed cet4.csv
var cet4CSV []byte

// Import 解析内嵌的 CET-4 CSV 并写入词库。
func Import(ctx context.Context, store *postgres.Store) (int, error) {
	reader := csv.NewReader(bytes.NewReader(cet4CSV))
	reader.FieldsPerRecord = 3

	header, err := reader.Read()
	if err != nil {
		return 0, err
	}
	if len(header) != 3 || header[0] != "word" || header[1] != "meaning" || header[2] != "level" {
		return 0, fmt.Errorf("expected CSV header word,meaning,level")
	}

	words := make([]domain.Word, 0, 4096)
	lineNumber := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
		lineNumber++
		level, err := strconv.Atoi(record[2])
		if err != nil {
			return 0, fmt.Errorf("line %d: invalid level %q", lineNumber, record[2])
		}
		word, err := domain.NewWord(record[0], record[1], level)
		if err != nil {
			return 0, fmt.Errorf("line %d: %w", lineNumber, err)
		}
		words = append(words, word)
	}
	return store.SeedWords(ctx, words)
}
