package records

import "testing"

func resetRecordsState(t *testing.T) {
	SetAll(nil)
	t.Cleanup(func() {
		SetAll(nil)
	})
}

func TestLookup(t *testing.T) {
	resetRecordsState(t)
	SetAll([]Record{{Name: "Example.COM", Type: TypeA, Value: "127.0.0.1"}})
	item, ok := Lookup1("example.com", TypeA)
	if !ok || item.Value != "127.0.0.1" || item.Name != "example.com." {
		t.Fatalf("got %v %+v", ok, item)
	}
	if !Delete("example.com", TypeA, "127.0.0.1") {
		t.Fatalf("delete failed")
	}
	if items := Lookup("example.com", TypeA); len(items) != 0 {
		t.Fatalf("expected empty, got %d", len(items))
	}
}
