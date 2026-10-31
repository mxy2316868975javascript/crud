package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mxy2316868975javascript/crud/internal/domain"
	"github.com/mxy2316868975javascript/crud/internal/httpapi"
	"github.com/mxy2316868975javascript/crud/internal/repository/postgres"
	"github.com/mxy2316868975javascript/crud/internal/service"
)

func testDatabaseURL() string {
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}
	return "postgres://postgres:postgres@localhost:5433/vocab_test?sslmode=disable"
}

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, testDatabaseURL())
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("postgres unavailable: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := postgres.EnsureSchema(ctx, pool); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}
	if _, err := pool.Exec(ctx, `TRUNCATE attempt_questions, attempts, words RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return pool
}

func newTestServer(t *testing.T) (*httptest.Server, *pgxpool.Pool) {
	t.Helper()
	pool := newTestPool(t)
	store := postgres.New(pool)
	mux := httpapi.NewMux(
		service.NewWordService(store),
		service.NewTestService(store),
		store,
	)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server, pool
}

func doJSON(t *testing.T, method, url string, body any) (*http.Response, []byte) {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request %s %s: %v", method, url, err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return response, responseBody
}

func decode[T any](t *testing.T, responseBody []byte) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(responseBody, &value); err != nil {
		t.Fatalf("decode %s: %v", responseBody, err)
	}
	return value
}

type wordInput struct {
	Word    string `json:"word"`
	Meaning string `json:"meaning"`
	Level   int    `json:"level"`
}

func createWord(t *testing.T, base string, word domain.Word) domain.Word {
	t.Helper()
	response, responseBody := doJSON(t, http.MethodPost, base+"/api/v1/words", wordInput{
		Word: word.Word, Meaning: word.Meaning, Level: word.Level,
	})
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create word: status %d, body %s", response.StatusCode, responseBody)
	}
	created := decode[domain.Word](t, responseBody)
	if created.ID == 0 {
		t.Fatalf("create word: missing id, body %s", responseBody)
	}
	if got, want := response.Header.Get("Location"), fmt.Sprintf("/api/v1/words/%d", created.ID); got != want {
		t.Fatalf("create word: Location %q, want %q", got, want)
	}
	return created
}

func TestHealth(t *testing.T) {
	server, _ := newTestServer(t)

	response, responseBody := doJSON(t, http.MethodGet, server.URL+"/healthz", nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("health: status %d, body %s", response.StatusCode, responseBody)
	}
	if !strings.Contains(string(responseBody), `"status":"ok"`) {
		t.Fatalf("health: unexpected body %s", responseBody)
	}
}

func TestHealthUnavailable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, testDatabaseURL())
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	pool.Close()

	store := postgres.New(pool)
	server := httptest.NewServer(httpapi.NewMux(
		service.NewWordService(store),
		service.NewTestService(store),
		store,
	))
	defer server.Close()

	response, responseBody := doJSON(t, http.MethodGet, server.URL+"/healthz", nil)
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("health: status %d, body %s", response.StatusCode, responseBody)
	}
}

func TestWordsCRUD(t *testing.T) {
	server, _ := newTestServer(t)
	base := server.URL

	created := createWord(t, base, domain.Word{Word: "abandon", Meaning: "v. 放弃", Level: 1})

	response, responseBody := doJSON(t, http.MethodPost, base+"/api/v1/words", wordInput{
		Word: "abandon", Meaning: "v. 放弃", Level: 1,
	})
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate word: status %d, body %s", response.StatusCode, responseBody)
	}

	response, responseBody = doJSON(t, http.MethodPost, base+"/api/v1/words", wordInput{
		Word: "invalid", Meaning: "x", Level: 6,
	})
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid level: status %d, body %s", response.StatusCode, responseBody)
	}

	response, responseBody = doJSON(t, http.MethodGet, base+"/api/v1/words?level=1&limit=10&offset=0", nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("list words: status %d, body %s", response.StatusCode, responseBody)
	}
	list := decode[struct {
		Items []domain.Word `json:"items"`
		Total int           `json:"total"`
	}](t, responseBody)
	if list.Total != 1 || len(list.Items) != 1 || list.Items[0].ID != created.ID {
		t.Fatalf("list words: unexpected body %s", responseBody)
	}

	response, responseBody = doJSON(t, http.MethodGet, fmt.Sprintf("%s/api/v1/words/%d", base, created.ID), nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("get word: status %d, body %s", response.StatusCode, responseBody)
	}

	response, responseBody = doJSON(t, http.MethodPut, fmt.Sprintf("%s/api/v1/words/%d", base, created.ID), wordInput{
		Word: "abandoned", Meaning: "adj. 被放弃的", Level: 2,
	})
	if response.StatusCode != http.StatusOK {
		t.Fatalf("put word: status %d, body %s", response.StatusCode, responseBody)
	}
	updated := decode[domain.Word](t, responseBody)
	if updated.Word != "abandoned" || updated.Meaning != "adj. 被放弃的" || updated.Level != 2 {
		t.Fatalf("put word: unexpected body %s", responseBody)
	}

	response, responseBody = doJSON(t, http.MethodPatch, fmt.Sprintf("%s/api/v1/words/%d", base, created.ID), map[string]int{"level": 5})
	if response.StatusCode != http.StatusOK {
		t.Fatalf("patch word: status %d, body %s", response.StatusCode, responseBody)
	}
	patched := decode[domain.Word](t, responseBody)
	if patched.Level != 5 || patched.Word != "abandoned" || patched.Meaning != "adj. 被放弃的" {
		t.Fatalf("patch word: unexpected body %s", responseBody)
	}

	response, responseBody = doJSON(t, http.MethodPatch, fmt.Sprintf("%s/api/v1/words/%d", base, created.ID), map[string]any{})
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("empty patch: status %d, body %s", response.StatusCode, responseBody)
	}

	response, responseBody = doJSON(t, http.MethodGet, base+"/api/v1/words/abc", nil)
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("bad id: status %d, body %s", response.StatusCode, responseBody)
	}

	response, responseBody = doJSON(t, http.MethodDelete, fmt.Sprintf("%s/api/v1/words/%d", base, created.ID), nil)
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("delete word: status %d, body %s", response.StatusCode, responseBody)
	}

	response, responseBody = doJSON(t, http.MethodGet, fmt.Sprintf("%s/api/v1/words/%d", base, created.ID), nil)
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("get deleted word: status %d, body %s", response.StatusCode, responseBody)
	}
}

func TestTestFlow(t *testing.T) {
	server, pool := newTestServer(t)
	base := server.URL

	for i := 0; i < 10; i++ {
		createWord(t, base, domain.Word{
			Word:    fmt.Sprintf("word%02d", i),
			Meaning: fmt.Sprintf("meaning %02d", i),
			Level:   i%5 + 1,
		})
	}

	response, responseBody := doJSON(t, http.MethodPost, base+"/api/v1/tests", map[string]int{"count": 5})
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create test: status %d, body %s", response.StatusCode, responseBody)
	}
	if strings.Contains(string(responseBody), "correct_index") {
		t.Fatalf("create test leaked correct_index: %s", responseBody)
	}
	test := decode[domain.Test](t, responseBody)
	if test.ID == 0 || test.QuestionCount != 5 || len(test.Questions) != 5 {
		t.Fatalf("create test: unexpected body %s", responseBody)
	}
	for _, q := range test.Questions {
		if q.WordID == 0 || q.Word == "" || len(q.Options) != 4 {
			t.Fatalf("create test: malformed question %+v", q)
		}
	}
	if got, want := response.Header.Get("Location"), fmt.Sprintf("/api/v1/tests/%d", test.ID); got != want {
		t.Fatalf("create test: Location %q, want %q", got, want)
	}

	response, responseBody = doJSON(t, http.MethodGet, fmt.Sprintf("%s/api/v1/tests/%d", base, test.ID), nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("get test: status %d, body %s", response.StatusCode, responseBody)
	}
	fetched := decode[domain.Test](t, responseBody)
	if fetched.Completed || len(fetched.Questions) != 5 {
		t.Fatalf("get test: unexpected body %s", responseBody)
	}

	response, responseBody = doJSON(t, http.MethodDelete, fmt.Sprintf("%s/api/v1/words/%d", base, test.Questions[0].WordID), nil)
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("delete used word: status %d, body %s", response.StatusCode, responseBody)
	}

	rows, err := pool.Query(context.Background(),
		`SELECT word_id, correct_index FROM attempt_questions WHERE attempt_id = $1`, test.ID)
	if err != nil {
		t.Fatalf("query correct indexes: %v", err)
	}
	answers := make([]domain.Answer, 0, 5)
	for rows.Next() {
		var answer domain.Answer
		if err := rows.Scan(&answer.WordID, &answer.SelectedIndex); err != nil {
			rows.Close()
			t.Fatalf("scan correct index: %v", err)
		}
		answers = append(answers, answer)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatalf("query correct indexes: %v", err)
	}
	if len(answers) != 5 {
		t.Fatalf("correct indexes: got %d, want 5", len(answers))
	}

	response, responseBody = doJSON(t, http.MethodPost, fmt.Sprintf("%s/api/v1/tests/%d/answers", base, test.ID), map[string]any{"answers": answers})
	if response.StatusCode != http.StatusOK {
		t.Fatalf("submit answers: status %d, body %s", response.StatusCode, responseBody)
	}
	result := decode[domain.Test](t, responseBody)
	if !result.Completed || result.CorrectCount == nil || *result.CorrectCount != 5 {
		t.Fatalf("submit answers: unexpected body %s", responseBody)
	}
	if result.EstimatedVocabulary == nil || *result.EstimatedVocabulary != 10 {
		t.Fatalf("submit answers: unexpected estimate %s", responseBody)
	}
	if result.CompletedAt == nil {
		t.Fatalf("submit answers: missing completed_at %s", responseBody)
	}

	response, responseBody = doJSON(t, http.MethodPost, fmt.Sprintf("%s/api/v1/tests/%d/answers", base, test.ID), map[string]any{"answers": answers})
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("resubmit answers: status %d, body %s", response.StatusCode, responseBody)
	}

	response, responseBody = doJSON(t, http.MethodGet, fmt.Sprintf("%s/api/v1/tests/%d", base, test.ID), nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("get completed test: status %d, body %s", response.StatusCode, responseBody)
	}
	completed := decode[domain.Test](t, responseBody)
	if !completed.Completed || completed.CorrectCount == nil || *completed.CorrectCount != 5 {
		t.Fatalf("get completed test: unexpected body %s", responseBody)
	}
}

func TestTestValidation(t *testing.T) {
	server, _ := newTestServer(t)
	base := server.URL

	response, responseBody := doJSON(t, http.MethodPost, base+"/api/v1/tests", map[string]int{"count": 5})
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("empty bank: status %d, body %s", response.StatusCode, responseBody)
	}

	for i := 0; i < 2; i++ {
		createWord(t, base, domain.Word{
			Word:    fmt.Sprintf("word%02d", i),
			Meaning: fmt.Sprintf("meaning %02d", i),
			Level:   1,
		})
	}
	response, responseBody = doJSON(t, http.MethodPost, base+"/api/v1/tests", map[string]int{"count": 5})
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("not enough words: status %d, body %s", response.StatusCode, responseBody)
	}

	for i := 2; i < 6; i++ {
		createWord(t, base, domain.Word{
			Word:    fmt.Sprintf("word%02d", i),
			Meaning: fmt.Sprintf("meaning %02d", i),
			Level:   1,
		})
	}
	response, responseBody = doJSON(t, http.MethodPost, base+"/api/v1/tests", map[string]int{"count": 2})
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create test: status %d, body %s", response.StatusCode, responseBody)
	}
	test := decode[domain.Test](t, responseBody)
	if len(test.Questions) != 2 {
		t.Fatalf("create test: unexpected body %s", responseBody)
	}
	id := test.ID

	response, responseBody = doJSON(t, http.MethodPost, fmt.Sprintf("%s/api/v1/tests/%d/answers", base, id), map[string]any{
		"answers": []domain.Answer{{WordID: test.Questions[0].WordID, SelectedIndex: 0}},
	})
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("missing answer: status %d, body %s", response.StatusCode, responseBody)
	}

	response, responseBody = doJSON(t, http.MethodPost, fmt.Sprintf("%s/api/v1/tests/%d/answers", base, id), map[string]any{
		"answers": []domain.Answer{
			{WordID: test.Questions[0].WordID, SelectedIndex: 0},
			{WordID: test.Questions[0].WordID, SelectedIndex: 1},
		},
	})
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("duplicate answer: status %d, body %s", response.StatusCode, responseBody)
	}

	response, responseBody = doJSON(t, http.MethodPost, fmt.Sprintf("%s/api/v1/tests/%d/answers", base, id), map[string]any{
		"answers": []domain.Answer{
			{WordID: test.Questions[0].WordID, SelectedIndex: 4},
			{WordID: test.Questions[1].WordID, SelectedIndex: 0},
		},
	})
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("out of range answer: status %d, body %s", response.StatusCode, responseBody)
	}

	response, responseBody = doJSON(t, http.MethodPost, fmt.Sprintf("%s/api/v1/tests/%d/answers", base, id), map[string]any{
		"answers": []domain.Answer{
			{WordID: test.Questions[0].WordID, SelectedIndex: 0},
			{WordID: 9999, SelectedIndex: 0},
		},
	})
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("unknown word answer: status %d, body %s", response.StatusCode, responseBody)
	}

	response, responseBody = doJSON(t, http.MethodPost, fmt.Sprintf("%s/api/v1/tests/%d/answers", base, id), map[string]any{
		"answers": []domain.Answer{
			{WordID: test.Questions[0].WordID, SelectedIndex: 0},
			{WordID: test.Questions[1].WordID, SelectedIndex: 0},
		},
	})
	if response.StatusCode != http.StatusOK {
		t.Fatalf("valid answers: status %d, body %s", response.StatusCode, responseBody)
	}
}

type schemaSpec struct {
	Properties map[string]struct {
		Nullable bool `json:"nullable"`
	} `json:"properties"`
}

func TestSwaggerDocs(t *testing.T) {
	server, _ := newTestServer(t)

	response, responseBody := doJSON(t, http.MethodGet, server.URL+"/swagger.json", nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("swagger: status %d, body %s", response.StatusCode, responseBody)
	}

	var swaggerSpec struct {
		Swagger string                                `json:"swagger"`
		Paths   map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(responseBody, &swaggerSpec); err != nil {
		t.Fatalf("swagger: invalid JSON: %v", err)
	}
	if swaggerSpec.Swagger != "2.0" {
		t.Fatalf("swagger: unexpected version %q", swaggerSpec.Swagger)
	}

	required := map[string][]string{
		"/healthz":                   {"get"},
		"/api/v1/words":              {"get", "post"},
		"/api/v1/words/{id}":         {"get", "put", "patch", "delete"},
		"/api/v1/tests":              {"post"},
		"/api/v1/tests/{id}":         {"get"},
		"/api/v1/tests/{id}/answers": {"post"},
	}
	for path, methods := range required {
		for _, method := range methods {
			if _, ok := swaggerSpec.Paths[path][method]; !ok {
				t.Errorf("swagger: missing %s %s", method, path)
			}
		}
	}

	response, responseBody = doJSON(t, http.MethodGet, server.URL+"/openapi.json", nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("openapi: status %d, body %s", response.StatusCode, responseBody)
	}
	var openAPISpec struct {
		OpenAPI    string                                `json:"openapi"`
		Paths      map[string]map[string]json.RawMessage `json:"paths"`
		Components struct {
			Schemas map[string]schemaSpec `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(responseBody, &openAPISpec); err != nil {
		t.Fatalf("openapi: invalid JSON: %v", err)
	}
	if openAPISpec.OpenAPI != "3.0.3" {
		t.Fatalf("openapi: unexpected version %q", openAPISpec.OpenAPI)
	}
	for path, methods := range required {
		for _, method := range methods {
			if _, ok := openAPISpec.Paths[path][method]; !ok {
				t.Errorf("openapi: missing %s %s", method, path)
			}
		}
	}

	foundNullable := false
	for name, schema := range openAPISpec.Components.Schemas {
		if name != "Test" && !strings.HasSuffix(name, ".Test") {
			continue
		}
		if schema.Properties["correct_count"].Nullable {
			foundNullable = true
		}
	}
	if !foundNullable {
		t.Errorf("openapi: Test.correct_count should be nullable")
	}

	response, responseBody = doJSON(t, http.MethodGet, server.URL+"/docs", nil)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("docs: status %d, body %s", response.StatusCode, responseBody)
	}
	if !strings.Contains(string(responseBody), "SwaggerUIBundle") || !strings.Contains(string(responseBody), "/openapi.json") {
		t.Fatalf("docs: unexpected body %s", responseBody)
	}
}
