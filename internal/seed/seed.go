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
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
		level, err := strconv.Atoi(record[2])
		if err != nil || level < 1 || level > 5 {
			return 0, fmt.Errorf("invalid level %q", record[2])
		}
		words = append(words, domain.Word{Word: record[0], Meaning: record[1], Level: level})
	}
	return store.SeedWords(ctx, words)
}
