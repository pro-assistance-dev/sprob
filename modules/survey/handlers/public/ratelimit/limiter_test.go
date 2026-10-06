package ratelimit

import (
	"testing"
	"time"
)

func TestAllowWithinLimit(t *testing.T) {
	l := New(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !l.Allow("ip1") {
			t.Fatalf("попытка %d должна проходить", i+1)
		}
	}
	if l.Allow("ip1") {
		t.Fatal("4-я попытка должна быть отклонена")
	}
	// Другой ключ — независимый счётчик.
	if !l.Allow("ip2") {
		t.Fatal("другой ключ не должен затрагиваться")
	}
}

func TestWindowExpires(t *testing.T) {
	l := New(1, 30*time.Millisecond)
	if !l.Allow("ip") {
		t.Fatal("первая попытка должна проходить")
	}
	if l.Allow("ip") {
		t.Fatal("вторая должна быть отклонена")
	}
	time.Sleep(40 * time.Millisecond)
	if !l.Allow("ip") {
		t.Fatal("после окна попытка должна снова проходить")
	}
}

func TestZeroLimitDisables(t *testing.T) {
	l := New(0, time.Minute)
	for i := 0; i < 100; i++ {
		if !l.Allow("ip") {
			t.Fatal("limit<=0 должен отключать лимит")
		}
	}
}
