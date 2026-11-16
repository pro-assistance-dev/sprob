package audit

import "testing"

func TestOperationFor(t *testing.T) {
	cases := []struct{ method, path string; payload map[string]interface{}; want string }{
		{"POST", "/api/worker-positions/dismiss", map[string]interface{}{"departmentId": "x"}, "update"},
		{"POST", "/api/worker-positions", map[string]interface{}{"workerId": "x"}, ""},
		{"POST", "/api/worker-positions", map[string]interface{}{"id": "x"}, "update"},
		{"POST", "/api/worker-documents/bulk-accept", map[string]interface{}{}, "update"},
		{"PUT", "/api/worker-positions/dismiss", map[string]interface{}{}, ""},
	}
	for _, c := range cases {
		if got := operationFor(c.method, c.path, c.payload); got != c.want {
			t.Errorf("operationFor(%s %s) = %q, want %q", c.method, c.path, got, c.want)
		}
	}
}
