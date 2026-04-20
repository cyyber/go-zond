package uint512

import (
	"math/big"
	"testing"
)

func mustHex(t *testing.T, s string) *big.Int {
	t.Helper()
	b, ok := new(big.Int).SetString(s, 0)
	if !ok {
		t.Fatalf("bad hex %q", s)
	}
	return b
}

func TestRoundTripBig(t *testing.T) {
	cases := []string{
		"0",
		"1",
		"0xdeadbeef",
		"0x" + "ff" + "00" + "11223344556677881122334455667788" +
			"11223344556677881122334455667788" + "1122334455667788112233445566",
	}
	for _, c := range cases {
		b := mustHex(t, c)
		z, overflow := FromBig(b)
		if overflow {
			t.Fatalf("unexpected overflow for %s", c)
		}
		if got := z.ToBig(); got.Cmp(b) != 0 {
			t.Fatalf("round-trip failed: want %x got %x", b, got)
		}
	}
}

func TestFromBigOverflow(t *testing.T) {
	b := new(big.Int).Lsh(big.NewInt(1), 513)
	_, overflow := FromBig(b)
	if !overflow {
		t.Fatal("expected overflow for 2^513")
	}
}

func TestAddWraps(t *testing.T) {
	a := new(Int).SetAllOne()
	b := NewInt(1)
	z := new(Int).Add(a, b)
	if !z.IsZero() {
		t.Fatalf("AllOne + 1 should wrap to 0, got %x", z.ToBig())
	}
}

func TestSubWraps(t *testing.T) {
	a := NewInt(0)
	b := NewInt(1)
	z := new(Int).Sub(a, b)
	// 0 - 1 (mod 2^512) = 2^512 - 1 = AllOne
	want := new(Int).SetAllOne()
	if !z.Eq(want) {
		t.Fatalf("0-1 should be 2^512-1")
	}
}

func TestMulWraps(t *testing.T) {
	// (2^256) * (2^256) = 2^512 ≡ 0 (mod 2^512)
	a := NewInt(1)
	a.Lsh(a, 256)
	z := new(Int).Mul(a, a)
	if !z.IsZero() {
		t.Fatalf("2^256 * 2^256 should be 0, got %x", z.ToBig())
	}
}

func TestBytes64RoundTrip(t *testing.T) {
	z := new(Int).SetAllOne() // 2^512 - 1
	b := z.Bytes64()
	for i, x := range b {
		if x != 0xff {
			t.Fatalf("byte %d: want 0xff got %x", i, x)
		}
	}
	y := new(Int).SetBytes(b[:])
	if !y.Eq(z) {
		t.Fatal("Bytes64 -> SetBytes mismatch")
	}
}

func TestByte(t *testing.T) {
	z := new(Int).SetAllOne()
	// byte at index 63 of AllOne == 0xff; result should be 0xff
	n := NewInt(63)
	z.Byte(n)
	if z.Uint64() != 0xff {
		t.Fatalf("Byte(63) want 0xff got %x", z.Uint64())
	}
}

func TestLshRsh(t *testing.T) {
	z := NewInt(1)
	z.Lsh(z, 511)
	// Should equal 2^511 — only the top bit of limb 7 set.
	want := new(Int)
	want[7] = 1 << 63
	if !z.Eq(want) {
		t.Fatal("Lsh to bit 511 mismatch")
	}
	z.Lsh(z, 1) // 2^512 masked to 0
	if !z.IsZero() {
		t.Fatal("Lsh past 512 should wrap to 0")
	}
}

func TestSignAndSgtSlt(t *testing.T) {
	neg := new(Int).SetAllOne() // -1 signed
	pos := NewInt(1)

	if neg.Sign() != -1 {
		t.Fatal("AllOne Sign should be -1")
	}
	if pos.Sign() != 1 {
		t.Fatal("1 Sign should be 1")
	}
	if !neg.Slt(pos) {
		t.Fatal("-1 Slt 1 should be true")
	}
	if !pos.Sgt(neg) {
		t.Fatal("1 Sgt -1 should be true")
	}
}

func TestSDivSMod(t *testing.T) {
	// -10 / 3 == -3, -10 % 3 == -1
	ten := NewInt(10)
	neg10 := new(Int).Sub(NewInt(0), ten)
	three := NewInt(3)

	q := new(Int).SDiv(neg10, three)
	qSigned := q.toSigned()
	if qSigned.Cmp(big.NewInt(-3)) != 0 {
		t.Fatalf("SDiv: want -3 got %s", qSigned)
	}

	r := new(Int).SMod(neg10, three)
	rSigned := r.toSigned()
	if rSigned.Cmp(big.NewInt(-1)) != 0 {
		t.Fatalf("SMod: want -1 got %s", rSigned)
	}
}

func TestExtendSign(t *testing.T) {
	// 0xff with byteNum=0 (one byte sign-extended) should become AllOne.
	x := NewInt(0xff)
	zero := NewInt(0)
	z := new(Int).ExtendSign(x, zero)
	allOne := new(Int).SetAllOne()
	if !z.Eq(allOne) {
		t.Fatalf("ExtendSign of 0xff byte 0: want AllOne got %x", z.ToBig())
	}

	// 0x7f with byteNum=0 stays 0x7f (sign bit clear).
	x2 := NewInt(0x7f)
	z2 := new(Int).ExtendSign(x2, zero)
	if z2.Uint64() != 0x7f {
		t.Fatalf("ExtendSign of 0x7f byte 0: want 0x7f got %x", z2.Uint64())
	}
}

func TestExpMod(t *testing.T) {
	// 2^512 mod 2^512 == 0 via Exp
	base := NewInt(2)
	exp := NewInt(512)
	z := new(Int).Exp(base, exp)
	if !z.IsZero() {
		t.Fatalf("2^512 mod 2^512 should be 0, got %x", z.ToBig())
	}
}

func TestBytes32TruncatesToLow256(t *testing.T) {
	z := new(Int).SetAllOne() // 2^512 - 1
	b32 := z.Bytes32()
	for i, x := range b32 {
		if x != 0xff {
			t.Fatalf("Bytes32 byte %d: want 0xff got %x", i, x)
		}
	}
}

func TestSRshNegative(t *testing.T) {
	// -1 >> 5 == -1 (sign-preserving)
	neg1 := new(Int).SetAllOne()
	z := new(Int).SRsh(neg1, 5)
	if !z.Eq(neg1) {
		t.Fatalf("SRsh of -1 should stay -1")
	}
}
