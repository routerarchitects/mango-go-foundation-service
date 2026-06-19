package routes_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/routerarchitects/mango-go-foundation-service/internal/http/routes"
	subsysteroutes "github.com/routerarchitects/ow-common-mods/fiber/system-routes"
)

func prettyJSON(raw []byte) string {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		return string(raw)
	}
	return pretty.String()
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

	t.Run("TC-LIVEZ-001", func(t *testing.T) {
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

func TestPrivateRoutesAuth(t *testing.T) {
	app := fiber.New()

	mockAuth := func(c fiber.Ctx) error {
		apiKey := c.Get("X-API-KEY")
		internalName := c.Get("X-INTERNAL-NAME")
		if apiKey == "expected-key" && internalName == "test-service" {
			return c.Next()
		}
		return c.SendStatus(http.StatusUnauthorized)
	}

	routes.RegisterPrivate(app, routes.PrivateDeps{
		AuthHandler: mockAuth,
		Subsystem:   subsysteroutes.Config{},
	})

	t.Run("TC-SYS-GET-005", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/system?command=info", nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("failed to test system info: %v", err)
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)

		fmt.Printf("\n------------------------------------------------------------\nTC-SYS-GET-005 (Retrieve System Diagnostics - Missing Auth Header)\nRequest: %s %s\nHeaders: %v\nResponse Status: %s\nResponse Body:\n%s\n",
			req.Method, req.URL.String(), req.Header, statusText(resp.StatusCode), string(bodyBytes))

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
		}
	})

	t.Run("TC-SYS-GET-001", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/system?command=info", nil)
		req.Header.Set("X-API-KEY", "expected-key")
		req.Header.Set("X-INTERNAL-NAME", "test-service")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("failed to test system info with credentials: %v", err)
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)

		fmt.Printf("\n------------------------------------------------------------\nTC-SYS-GET-001 (Retrieve System Diagnostics - Valid Auth Header)\nRequest: %s %s\nHeaders: %v\nResponse Status: %s\nResponse Body:\n%s\n",
			req.Method, req.URL.String(), req.Header, statusText(resp.StatusCode), prettyJSON(bodyBytes))

		if resp.StatusCode == http.StatusUnauthorized {
			t.Errorf("expected request to pass authentication, but got 401")
		}
	})

	t.Run("TC-SYS-POST-001", func(t *testing.T) {
		requestBody := `{"command":"setloglevel","subsystems":[{"tag":"http","value":"debug"}]}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/system", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-KEY", "expected-key")
		req.Header.Set("X-INTERNAL-NAME", "test-service")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("failed to test post system with credentials: %v", err)
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)

		fmt.Printf("\n------------------------------------------------------------\nTC-SYS-POST-001 (Set log level successfully for a subsystem)\nRequest: %s %s\nHeaders: %v\nRequest Body:\n%s\nResponse Status: %s\nResponse Body:\n%s\n",
			req.Method, req.URL.String(), req.Header, prettyJSON([]byte(requestBody)), statusText(resp.StatusCode), prettyJSON(bodyBytes))

		if resp.StatusCode == http.StatusUnauthorized {
			t.Errorf("expected request to pass authentication, but got 401")
		}
	})
}
