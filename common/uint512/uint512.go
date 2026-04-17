// Package uint512 provides a 512-bit unsigned integer type with an API
// compatible with github.com/holiman/uint256.Int.
//
// The current implementation backs each Int with math/big and masks every
// arithmetic result to 512 bits. The API-compatible surface is intentionally
// chosen so this implementation can later be swapped for a fork of
// holiman/uint256 (or a dedicated [8]uint64 implementation) without touching
// any call site.
package uint512

import (
	"fmt"
	"math/big"
)

// Int is a 512-bit unsigned integer.
//
// The zero value is ready to use and represents 0. Int values must not be
// copied while holding a reference to their internal state; use Set or Clone
// for semantic copies.
type Int struct {
	v big.Int
}

// modulus = 2^512; maskAll = 2^512 - 1.
var (
	modulus = new(big.Int).Lsh(big.NewInt(1), 512)
	maskAll = new(big.Int).Sub(modulus, big.NewInt(1))

	// signBit = 2^511 (MSB of a 512-bit value).
	signBit = new(big.Int).Lsh(big.NewInt(1), 511)
)

// mask reduces v modulo 2^512 in place.
func (z *Int) mask() *Int {
	z.v.And(&z.v, maskAll)
	return z
}

// NewInt returns a new Int with value v.
func NewInt(v uint64) *Int {
	z := &Int{}
	z.v.SetUint64(v)
	return z
}

// FromBig returns an Int with value b and a boolean indicating whether the
// value fit in 512 bits (without wrapping).
func FromBig(b *big.Int) (*Int, bool) {
	z := &Int{}
	overflow := z.SetFromBig(b)
	return z, overflow
}

// MustFromBig is like FromBig but panics on overflow.
func MustFromBig(b *big.Int) *Int {
	z, overflow := FromBig(b)
	if overflow {
		panic("uint512: value overflows 512 bits")
	}
	return z
}

// --- basic accessors ---------------------------------------------------------

// Uint64 returns the low 64 bits of z.
func (z *Int) Uint64() uint64 {
	return z.v.Uint64()
}

// Uint64WithOverflow returns the low 64 bits of z together with a flag
// indicating whether z exceeds 64 bits.
func (z *Int) Uint64WithOverflow() (uint64, bool) {
	if z.v.BitLen() > 64 {
		return z.v.Uint64(), true
	}
	return z.v.Uint64(), false
}

// IsUint64 reports whether z fits in a uint64.
func (z *Int) IsUint64() bool {
	return z.v.BitLen() <= 64
}

// IsZero reports whether z is 0.
func (z *Int) IsZero() bool {
	return z.v.Sign() == 0
}

// Sign returns -1, 0 or 1 interpreting z as a signed 512-bit integer.
// An unsigned zero returns 0; values with the MSB set are negative.
func (z *Int) Sign() int {
	if z.v.Sign() == 0 {
		return 0
	}
	if z.v.Cmp(signBit) >= 0 {
		return -1
	}
	return 1
}

// BitLen returns the number of bits required to represent z.
func (z *Int) BitLen() int {
	return z.v.BitLen()
}

// Cmp compares z and x as unsigned integers and returns -1, 0 or 1.
func (z *Int) Cmp(x *Int) int {
	return z.v.Cmp(&x.v)
}

// Eq reports whether z == x.
func (z *Int) Eq(x *Int) bool {
	return z.v.Cmp(&x.v) == 0
}

// Lt reports whether z < x (unsigned).
func (z *Int) Lt(x *Int) bool {
	return z.v.Cmp(&x.v) < 0
}

// Gt reports whether z > x (unsigned).
func (z *Int) Gt(x *Int) bool {
	return z.v.Cmp(&x.v) > 0
}

// LtUint64 reports whether z < u (unsigned).
func (z *Int) LtUint64(u uint64) bool {
	if z.v.BitLen() > 64 {
		return false
	}
	return z.v.Uint64() < u
}

// GtUint64 reports whether z > u (unsigned).
func (z *Int) GtUint64(u uint64) bool {
	if z.v.BitLen() > 64 {
		return true
	}
	return z.v.Uint64() > u
}

// toSigned returns z interpreted as a signed 512-bit integer in a fresh big.Int.
func (z *Int) toSigned() *big.Int {
	if z.Sign() < 0 {
		return new(big.Int).Sub(&z.v, modulus)
	}
	return new(big.Int).Set(&z.v)
}

// Slt reports whether z < x when both are interpreted as signed 512-bit.
func (z *Int) Slt(x *Int) bool {
	return z.toSigned().Cmp(x.toSigned()) < 0
}

// Sgt reports whether z > x when both are interpreted as signed 512-bit.
func (z *Int) Sgt(x *Int) bool {
	return z.toSigned().Cmp(x.toSigned()) > 0
}

// --- setters -----------------------------------------------------------------

// Set assigns x to z and returns z.
func (z *Int) Set(x *Int) *Int {
	z.v.Set(&x.v)
	return z
}

// SetUint64 sets z to v and returns z.
func (z *Int) SetUint64(v uint64) *Int {
	z.v.SetUint64(v)
	return z
}

// SetBytes interprets buf as the big-endian bytes of an unsigned integer,
// truncates to 512 bits if longer, and sets z. Returns z.
func (z *Int) SetBytes(buf []byte) *Int {
	if len(buf) > 64 {
		buf = buf[len(buf)-64:]
	}
	z.v.SetBytes(buf)
	return z
}

// SetFromBig sets z to b mod 2^512 and reports whether b overflowed
// (was outside [0, 2^512)).
func (z *Int) SetFromBig(b *big.Int) bool {
	z.v.Set(b)
	overflow := z.v.Sign() < 0 || z.v.Cmp(modulus) >= 0
	z.mask()
	return overflow
}

// SetAllOne sets z to 2^512 - 1 and returns z.
func (z *Int) SetAllOne() *Int {
	z.v.Set(maskAll)
	return z
}

// SetOne sets z to 1 and returns z.
func (z *Int) SetOne() *Int {
	z.v.SetUint64(1)
	return z
}

// Clear sets z to 0 and returns z.
func (z *Int) Clear() *Int {
	z.v.SetUint64(0)
	return z
}

// Reset is an alias for Clear kept for API compatibility.
func (z *Int) Reset() {
	z.Clear()
}

// Clone returns a deep copy of z.
func (z *Int) Clone() *Int {
	return new(Int).Set(z)
}

// ToBig returns z as a fresh *big.Int.
func (z *Int) ToBig() *big.Int {
	return new(big.Int).Set(&z.v)
}

// --- byte serialization ------------------------------------------------------

// Bytes returns the minimal big-endian byte representation of z (no leading
// zeros). An empty slice is returned for zero.
func (z *Int) Bytes() []byte {
	return z.v.Bytes()
}

// Bytes32 returns the 32-byte big-endian representation of the low 256 bits of z.
func (z *Int) Bytes32() [32]byte {
	var out [32]byte
	b := z.v.Bytes()
	if len(b) > 32 {
		b = b[len(b)-32:]
	}
	copy(out[32-len(b):], b)
	return out
}

// Hex returns the hexadecimal representation of z prefixed with "0x".
func (z *Int) Hex() string {
	return "0x" + z.v.Text(16)
}

// Format implements fmt.Formatter. It supports the same verbs as big.Int.
func (z Int) Format(s fmt.State, verb rune) {
	z.v.Format(s, verb)
}

// Bytes64 returns the 64-byte big-endian representation of z.
func (z *Int) Bytes64() [64]byte {
	var out [64]byte
	b := z.v.Bytes()
	if len(b) > 64 {
		b = b[len(b)-64:]
	}
	copy(out[64-len(b):], b)
	return out
}

// --- arithmetic --------------------------------------------------------------

// Add sets z = x + y (mod 2^512) and returns z.
func (z *Int) Add(x, y *Int) *Int {
	z.v.Add(&x.v, &y.v)
	return z.mask()
}

// Sub sets z = x - y (mod 2^512) and returns z.
func (z *Int) Sub(x, y *Int) *Int {
	z.v.Sub(&x.v, &y.v)
	if z.v.Sign() < 0 {
		z.v.Add(&z.v, modulus)
	}
	return z.mask()
}

// Mul sets z = x * y (mod 2^512) and returns z.
func (z *Int) Mul(x, y *Int) *Int {
	z.v.Mul(&x.v, &y.v)
	return z.mask()
}

// Div sets z = x / y (unsigned, truncated). If y == 0 the result is 0.
func (z *Int) Div(x, y *Int) *Int {
	if y.v.Sign() == 0 {
		z.v.SetUint64(0)
		return z
	}
	z.v.Quo(&x.v, &y.v)
	return z.mask()
}

// Mod sets z = x mod y (unsigned). If y == 0 the result is 0.
func (z *Int) Mod(x, y *Int) *Int {
	if y.v.Sign() == 0 {
		z.v.SetUint64(0)
		return z
	}
	z.v.Rem(&x.v, &y.v)
	return z.mask()
}

// SDiv sets z = x / y where x and y are interpreted as signed 512-bit integers.
// Division rounds toward zero (EVM semantics). If y == 0 the result is 0.
func (z *Int) SDiv(x, y *Int) *Int {
	if y.v.Sign() == 0 {
		z.v.SetUint64(0)
		return z
	}
	sx := x.toSigned()
	sy := y.toSigned()
	res := new(big.Int).Quo(sx, sy)
	if res.Sign() < 0 {
		res.Add(res, modulus)
	}
	z.v.Set(res)
	return z.mask()
}

// SMod sets z = x mod y with EVM signed semantics (result takes the sign of x).
// If y == 0 the result is 0.
func (z *Int) SMod(x, y *Int) *Int {
	if y.v.Sign() == 0 {
		z.v.SetUint64(0)
		return z
	}
	sx := x.toSigned()
	sy := y.toSigned()
	res := new(big.Int).Rem(sx, sy)
	if res.Sign() < 0 {
		res.Add(res, modulus)
	}
	z.v.Set(res)
	return z.mask()
}

// Exp sets z = base ** exp (mod 2^512) and returns z.
func (z *Int) Exp(base, exp *Int) *Int {
	z.v.Exp(&base.v, &exp.v, modulus)
	return z
}

// AddMod sets z = (x + y) mod m. If m == 0 the result is 0.
func (z *Int) AddMod(x, y, m *Int) *Int {
	if m.v.Sign() == 0 {
		z.v.SetUint64(0)
		return z
	}
	sum := new(big.Int).Add(&x.v, &y.v)
	z.v.Rem(sum, &m.v)
	return z.mask()
}

// MulMod sets z = (x * y) mod m. If m == 0 the result is 0.
func (z *Int) MulMod(x, y, m *Int) *Int {
	if m.v.Sign() == 0 {
		z.v.SetUint64(0)
		return z
	}
	prod := new(big.Int).Mul(&x.v, &y.v)
	z.v.Rem(prod, &m.v)
	return z.mask()
}

// ExtendSign sets z to x sign-extended from byte position (byteNum+1) to 512 bits.
// Matches the semantics of EVM SIGNEXTEND but for 64-byte words.
func (z *Int) ExtendSign(x, byteNum *Int) *Int {
	if !byteNum.IsUint64() || byteNum.Uint64() >= 63 {
		z.Set(x)
		return z
	}
	bn := byteNum.Uint64()
	// Bit position of the sign bit within x.
	bit := uint(bn)*8 + 7
	mask := new(big.Int).Lsh(big.NewInt(1), bit+1)
	mask.Sub(mask, big.NewInt(1))
	low := new(big.Int).And(&x.v, mask)

	signMask := new(big.Int).Lsh(big.NewInt(1), bit)
	if new(big.Int).And(&x.v, signMask).Sign() != 0 {
		// Sign bit set: fill high bits with 1.
		high := new(big.Int).Xor(maskAll, mask)
		z.v.Or(low, high)
	} else {
		z.v.Set(low)
	}
	return z.mask()
}

// --- bitwise -----------------------------------------------------------------

// And sets z = x & y and returns z.
func (z *Int) And(x, y *Int) *Int {
	z.v.And(&x.v, &y.v)
	return z
}

// Or sets z = x | y and returns z.
func (z *Int) Or(x, y *Int) *Int {
	z.v.Or(&x.v, &y.v)
	return z.mask()
}

// Xor sets z = x ^ y and returns z.
func (z *Int) Xor(x, y *Int) *Int {
	z.v.Xor(&x.v, &y.v)
	return z.mask()
}

// Not sets z = ^x (bitwise complement within 512 bits) and returns z.
func (z *Int) Not(x *Int) *Int {
	z.v.Xor(&x.v, maskAll)
	return z
}

// Lsh sets z = x << n (mod 2^512) and returns z.
func (z *Int) Lsh(x *Int, n uint) *Int {
	z.v.Lsh(&x.v, n)
	return z.mask()
}

// Rsh sets z = x >> n (logical shift) and returns z.
func (z *Int) Rsh(x *Int, n uint) *Int {
	z.v.Rsh(&x.v, n)
	return z
}

// SRsh sets z = x >> n with sign extension (x interpreted as signed 512-bit).
func (z *Int) SRsh(x *Int, n uint) *Int {
	if n == 0 {
		return z.Set(x)
	}
	if x.Sign() >= 0 {
		z.v.Rsh(&x.v, n)
		return z
	}
	// Signed: shift the two's-complement negative representation.
	s := x.toSigned() // negative big.Int
	s.Rsh(s, n)      // big.Int.Rsh is arithmetic for negatives
	if s.Sign() < 0 {
		s.Add(s, modulus)
	}
	z.v.Set(s)
	return z.mask()
}

// Byte replaces z with the byte at index n (big-endian) of itself, where index 0
// is the most significant byte of the 64-byte representation. If n >= 64, z is set to 0.
func (z *Int) Byte(n *Int) *Int {
	if !n.IsUint64() {
		z.v.SetUint64(0)
		return z
	}
	idx := n.Uint64()
	if idx >= 64 {
		z.v.SetUint64(0)
		return z
	}
	// byte at big-endian index idx corresponds to bits [8*(63-idx), 8*(63-idx)+8)
	shift := uint(8 * (63 - idx))
	tmp := new(big.Int).Rsh(&z.v, shift)
	tmp.And(tmp, big.NewInt(0xff))
	z.v.Set(tmp)
	return z
}
