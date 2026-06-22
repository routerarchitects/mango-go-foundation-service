package routes_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/routerarchitects/mango-go-foundation-service/internal/http/routes"
	subsysteroutes "github.com/routerarchitects/ow-common-mods/fiber/system-routes"
)

type testResult struct {
	ID     string
	Desc   string
	Status string
}

var (
	resultsMu sync.Mutex
	results   []testResult
)

func recordResult(id, desc, status string) {
	resultsMu.Lock()
	results = append(results, testResult{ID: id, Desc: desc, Status: status})
	resultsMu.Unlock()
}

func tRun(t *testing.T, name string, desc string, fn func(t *testing.T)) {
	t.Run(name, func(t *testing.T) {
		defer func() {
			status := "PASS"
			if t.Failed() {
				status = "FAIL"
			} else if t.Skipped() {
				status = "SKIP"
			}
			recordResult(t.Name(), desc, status)
		}()
		fn(t)
	})
}

func TestMain(m *testing.M) {
	code := m.Run()
	printSummaryTable()
	os.Exit(code)
}

func printSummaryTable() {
	resultsMu.Lock()
	defer resultsMu.Unlock()
	if len(results) == 0 {
		return
	}

	fmt.Println("\n=====================================================================================================================================")
	fmt.Println("                                                         TEST SUMMARY REPORT                                                         ")
	fmt.Println("=====================================================================================================================================")
	fmt.Printf("%-4s | %-35s | %-70s | %-15s\n", "S.No", "TestCase", "Name", "Result")
	fmt.Println("-------------------------------------------------------------------------------------------------------------------------------------")
	for i, r := range results {
		color := "32"
		if r.Status == "FAIL" {
			color = "31"
		} else if r.Status == "SKIP" {
			color = "33"
		}
		statusStr := fmt.Sprintf("\033[%sm%s\033[0m", color, r.Status)

		displayName := r.ID
		if parts := strings.Split(displayName, "/"); len(parts) > 1 {
			displayName = parts[len(parts)-1]
		}
		fmt.Printf("%-4d | %-35s | %-70s | %-15s\n", i+1, displayName, r.Desc, statusStr)
	}
	fmt.Println("=====================================================================================================================================")
}

func statusText(code int) string {
	return fmt.Sprintf("%d %s", code, http.StatusText(code))
}

func TestRoutesLiveness(t *testing.T) {
	app := fiber.New()

	mockAuth := func(c fiber.Ctx) error {
		return c.Next()
	}

	routes.RegisterPublic(app, routes.PublicDeps{
		AuthHandler: mockAuth,
		Subsystem:   subsysteroutes.Config{},
	})

	tRun(t, "TC-LIVEZ-001", "Liveness probe returns 200 OK", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/livez", nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("failed to test liveness probe: %v", err)
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)

		fmt.Printf("\n------------------------------------------------------------\nTC-LIVEZ-001 (Liveness probe returns 200 OK)\nRequest: %s %s\nHeaders: %v\nResponse Status: %s\nResponse Body:\n%s\n",
			req.Method, req.URL.String(), req.Header, statusText(resp.StatusCode), string(bodyBytes))

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})
}
