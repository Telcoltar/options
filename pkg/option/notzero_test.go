package option

import "testing"

func TestBase_NotZero_Int(t *testing.T) {
	opt := NewBase[int]("int")
	if opt.NotZero() {
		t.Error("expected NotZero to be false when no value set")
	}

	opt.Default(0)
	if opt.NotZero() {
		t.Error("expected NotZero to be false for default 0")
	}

	opt.Default(5)
	if !opt.NotZero() {
		t.Error("expected NotZero to be true for default 5")
	}

	opt.Set(0)
	if opt.NotZero() {
		t.Error("expected NotZero to be false for set 0")
	}

	opt.Set(7)
	if !opt.NotZero() {
		t.Error("expected NotZero to be true for set 7")
	}
}

func TestBase_NotZero_String(t *testing.T) {
	opt := NewBase[string]("str")
	if opt.NotZero() {
		t.Error("expected NotZero to be false when no value set")
	}

	opt.Default("")
	if opt.NotZero() {
		t.Error("expected NotZero to be false for default empty string")
	}

	opt.Default("x")
	if !opt.NotZero() {
		t.Error("expected NotZero to be true for default 'x'")
	}
}

func TestBase_NotZero_Bool(t *testing.T) {
	opt := NewBase[bool]("bool")
	opt.Default(false)
	if opt.NotZero() {
		t.Error("expected NotZero to be false for default false")
	}
	opt.Set(true)
	if !opt.NotZero() {
		t.Error("expected NotZero to be true for true")
	}
}

func TestSlice_NotZero(t *testing.T) {
	opt := NewSlice[int]("slice")
	if opt.NotZero() {
		t.Error("expected NotZero to be false when no value set")
	}
	opt.EmptyDefault() // empty slice (non-nil) != zero (nil)
	if !opt.NotZero() {
		t.Error("expected NotZero to be true for empty default slice")
	}
	opt.Set([]int{})
	if !opt.NotZero() {
		t.Error("expected NotZero to be true for empty but non-nil slice value")
	}
}

func TestMap_NotZero(t *testing.T) {
	opt := NewMap[int]("map")
	if opt.NotZero() {
		t.Error("expected NotZero to be false when no value set")
	}
	opt.EmptyDefault() // empty map (non-nil) != zero (nil)
	if !opt.NotZero() {
		t.Error("expected NotZero to be true for empty default map")
	}
	opt.Set(map[string]int{})
	if !opt.NotZero() {
		t.Error("expected NotZero to be true for empty but non-nil map value")
	}
}
