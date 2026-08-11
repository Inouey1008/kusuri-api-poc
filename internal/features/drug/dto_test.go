package drug

import "testing"

func TestDrug_ToResponse(t *testing.T) {
	d := Drug{YJCode: "2189018F1043", Name: "エゼチミブ錠10mg「JG」"}

	got := d.toResponse()

	if got.YJCode != d.YJCode {
		t.Errorf("YJCode = %q, want %q", got.YJCode, d.YJCode)
	}
	if got.Name != d.Name {
		t.Errorf("Name = %q, want %q", got.Name, d.Name)
	}
}
