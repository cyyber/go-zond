# Architektura go-qrl — moduły istotne dla migracji 48B + 512-bit VM

> Dokumentacja techniczna skierowana do zaawansowanego programisty, który nie zna ani go-qrl, ani szerzej blockchainów w stylu Ethereum. Celem jest dać czytelnikowi model mentalny wystarczający do czytania i modyfikowania modułów objętych migracją.
>
> Dokument uzupełniający: [`ADR-ANALYSIS.md`](./ADR-ANALYSIS.md) (decyzje) i [`PRODUCTION-PLAN.md`](./PRODUCTION-PLAN.md) (plan pracy).

---

## 1. Kontekst projektu

### Czym jest go-qrl

**go-qrl** to pełny węzeł blockchain QRL (Quantum Resistant Ledger) napisany w Go. Wywodzi się bezpośrednio z **go-ethereum** (pierwotnego klienta Ethereum), z zastąpieniem:
- Kryptografii secp256k1 ECDSA → **ML-DSA-87** (post-kwantowy, NIST-standaryzowany).
- Szerokości adresu (20B → 48B).
- Szerokości słowa VM (256-bit → 512-bit).
- Nazw (EVM → QRVM, geth → gqrl, evmc → qrvmc, evmone → qrvmone).

Intencja biznesowa: post-kwantowy blockchain programowalny, z dziedziczoną od Ethereum programowalnością (smart-kontrakty, ERC-like tokeny, itp.).

### Co robi węzeł blockchain

Węzeł ma 4 podstawowe zadania:

1. **Sieciowe** — utrzymywać połączenia peer-to-peer z innymi węzłami, propagować transakcje i bloki.
2. **Konsensus** — weryfikować bloki i uzgadniać globalny porządek transakcji. W QRL: **Beacon chain / Proof-of-Stake** (ale to poza zakresem tego dokumentu).
3. **Wykonanie** — dla każdej transakcji odtworzyć deterministyczny skutek: zmiany kont, deploy kontraktów, emisje logów. To robi **QRVM** (wirtualna maszyna).
4. **Przechowywanie** — persystować łańcuch bloków, stan wszystkich kont, historię logów.

Plus dodatki: **JSON-RPC** dla klientów (portfele, dev tools), **GraphQL**, **signer** (oddzielny serwis do podpisywania).

### Model wykonania: "świat" to czysta funkcja transakcji

Kluczowy model mentalny:

```
stan_po = apply(blok, stan_przed)
```

Blockchain to **deterministyczny automat** — dane wejściowe (blok z podpisanymi transakcjami) i stan przed dają jednoznaczny stan po. Każdy uczciwy węzeł wykonuje tą samą funkcję i dochodzi do tego samego wyniku. Rozbieżność = bug albo złośliwy kod.

To dlatego tyle miejsca poświęcamy determinizmowi:
- Hasze muszą się zgadzać.
- Gas koszty muszą być identyczne na każdej maszynie.
- Porządek operacji jest ściśle zdefiniowany.
- Niedeterministyczne konstrukty (np. iteracja po Go map) są zabronione w hot path.

---

## 2. Warstwowy model architektury

Od góry (user-facing) do dołu (dysk):

```
┌────────────────────────────────────────────────────────────────────┐
│  JSON-RPC / GraphQL / WebSocket    (internal/qrlapi, graphql)      │ ← klienci zewn.
├────────────────────────────────────────────────────────────────────┤
│  Wallet / Signer / Accounts        (accounts/*, signer/*)          │ ← zarządzanie kluczami
├────────────────────────────────────────────────────────────────────┤
│  ABI / Bindings                    (accounts/abi)                  │ ← interop Go ↔ kontrakty
├────────────────────────────────────────────────────────────────────┤
│  QRVM (interpreter, state, types)  (core/vm, core/state,           │ ← wykonanie
│                                    core/types, core/*.go)          │
├────────────────────────────────────────────────────────────────────┤
│  Transaction pool                  (core/txpool)                   │ ← pending txs
├────────────────────────────────────────────────────────────────────┤
│  Downloader / Sync / Protocol      (qrl/downloader, qrl/protocols) │ ← synchronizacja
├────────────────────────────────────────────────────────────────────┤
│  P2P transport                     (p2p/*)                         │ ← sieć
├────────────────────────────────────────────────────────────────────┤
│  Trie + State DB                   (trie/*, core/state)            │ ← Merkle-Patricia
├────────────────────────────────────────────────────────────────────┤
│  Raw DB (LevelDB/Pebble)           (core/rawdb, qrldb)             │ ← key-value na dysku
└────────────────────────────────────────────────────────────────────┘
```

**Migracja 48B/512-bit dotyka warstwy wykonania (QRVM + state + types), ABI, i trochę API.** Warstwy poniżej (trie, DB) i powyżej (RPC) są w większości automatyczne, bo używają typu `common.Address`.

---

## 3. Podstawowe typy

### `common.Address`

Plik: `common/types.go`

```go
const AddressLength = 48  // było 20

type Address [AddressLength]byte
```

**Przed migracją:** `[20]byte`. Reprezentuje 160-bitowy identyfikator konta lub kontraktu.
**Po migracji:** `[48]byte`. Reprezentuje 384-bitowy identyfikator.

Kluczowe operacje:
- `BytesToAddress(b []byte) Address` — kopiuje bajty, crop z lewej jeśli input > 48B, padding zerami z lewej jeśli input < 48B.
- `NewAddressFromString(s string) (Address, error)` — parsuje `"Q" + 96 hex znaków`.
- `Hex() string` — zwraca kanoniczną reprezentację `Q` + 96 małych znaków hex. Adres QRL nie zawiera mixed-case checksumu.
- `Bytes() []byte` — surowe 48 bajtów.

Format tekstowy: `Q` + 96 hex. Przykład:
```
Q000000000000000000000000000000000000000000000000000000005aaeb6053f3e94c9b9a09f33669435e7ef1beaed
```

### `common.Hash`

```go
const HashLength = 32
type Hash [HashLength]byte
```

**Nie zmieniono w migracji.** Reprezentuje 256-bitowy hash Keccak-256. Używane wszędzie dla:
- Block hash
- Transaction hash
- State root, transactions root, receipts root
- Storage slot key i value
- Log topic
- Code hash

Jeśli ADR-001/002 zdecyduje poszerzyć storage/topic, `common.Hash` zostanie, ale pojawi się nowy typ (np. `common.StorageValue` = `[64]byte`).

### `uint256.Int` (holiman) i nasz `uint512.Int`

EVM operuje na 256-bitowych liczbach całkowitych. Biblioteka `github.com/holiman/uint256` dostarcza fiksowaną reprezentację `[4]uint64` z performancyjnymi metodami. Używana jako **słowo stosu** i w wielu miejscach arytmetyki.

Po migracji VM używa 512-bitowego słowa. Stworzyliśmy własny pakiet `common/uint512` z **identycznym API** jak holiman/uint256, ale 512-bitowy.

**Obecnie:** wrapper na `math/big.Int`. Działa, ale ~5-10× wolniejszy.
**Docelowo** (Faza 1.2): fork holimana albo `[8]uint64`.

---

## 4. Moduły — deep dive

### 4.1 `common/` — typy podstawowe

**Pliki:**
- `types.go` — `Address`, `Hash`, ich metody (`Hex`, `Bytes`, `String`, `UnmarshalJSON`, etc.).
- `big.go` — stałe `big.Int`: `Big0, Big1, Big32, Big64, Big256`.
- `bytes.go` — pomocnicze na bajtach: `LeftPadBytes`, `RightPadBytes`, `FromHex`, `Hex2Bytes`.
- `math/big.go` — wrappery wokół `big.Int`: `PaddedBigBytes`, `U256`, `U256Bytes`, **`U512`, `U512Bytes`** (nasze dodatki).

**Jak Address jest formatowany:**

```go
// hex() zwraca "Q" + lowercase hex (97 chars)
func (a Address) hex() []byte {
    var buf [len(a)*2 + 1]byte  // 48*2 + 1 = 97
    copy(buf[:1], hexutil.PrefixQ)   // 'Q'
    hex.Encode(buf[1:], a[:])
    return buf[:]
}

func (a Address) Hex() string {
    return string(a.hex())
}
```

Adres tekstowy nie używa Keccak ani mixed-case checksumu. Walidacja sprawdza strukturę: prefiks `Q` i 96 znaków hex.

### 4.2 `crypto/` — derivacja adresów, hashe, podpisy

**Pliki:**
- `crypto.go` — derivacja adresów, hashe Keccak, importy kluczy ECDSA (legacy).
- `signature_cgo.go`, `signature_nocgo.go` — podpisy secp256k1 (legacy).
- `pqcrypto/wallet/wallet.go` — wrapper nad **go-qrllib** (produkcyjny wallet ML-DSA-87).
- `crypto_test.go` — testy.

**Kluczowe funkcje w `crypto.go`:**

```go
// Pomocnicza: 48-bajtowy adres z dowolnych bajtów.
// Obecna implementacja (Opcja A w ADR-004):
func keccakToAddress48(data ...[]byte) common.Address {
    h1 := Keccak256(data...)     // 32 B
    h2 := Keccak256(h1)          // 32 B
    return common.BytesToAddress(append(h1, h2[:16]...))  // 48 B
}

// Adres kontraktu dla CREATE (zależy od nonce):
func CreateAddress(sender common.Address, nonce uint64) common.Address {
    data, _ := rlp.EncodeToBytes([]any{sender, nonce})
    return keccakToAddress48(data)
}

// Adres kontraktu dla CREATE2 (deterministyczny):
func CreateAddress2(sender common.Address, salt [32]byte, inithash []byte) common.Address {
    return keccakToAddress48([]byte{0xff}, sender.Bytes(), salt[:], inithash)
}

// Adres z ECDSA pubkey (legacy/test):
func PubkeyToAddress(p ecdsa.PublicKey) common.Address {
    pubBytes := FromECDSAPub(&p)
    return keccakToAddress48(pubBytes[1:])  // pomijamy 0x04 prefix SEC1
}
```

**Keccak-256 (Legacy Keccak, nie SHA-3):**

```go
func Keccak256(data ...[]byte) []byte {
    b := make([]byte, 32)
    d := sha3.NewLegacyKeccak256()
    for _, b := range data {
        d.Write(b)
    }
    d.Sum(b[:0])
    return b
}
```

`sha3.NewLegacyKeccak256` to Keccak pre-FIPS-202 padding (który Ethereum użyło jako pierwszy), różniący się 1-bitem od sha3.NewSum256 (FIPS-202). **Nie mylić!**

**Wallet path (produkcyjny):**

```go
// crypto/pqcrypto/wallet/wallet.go
func (w *MLDSA87Wallet) GetAddress() [common.AddressLength]uint8 {
    // go-qrllib generates native 48-byte QRL addresses.
    oldAddr := w.Wallet.GetAddress()
    var addr [common.AddressLength]uint8
    copy(addr[common.AddressLength-len(oldAddr):], oldAddr[:])
    return addr
}
```

**⚠️ Ostrzeżenie:** obecnie `GetAddress()` zwraca **20-bajtowy adres z qrllib uzupełniony zerami do 48 bajtów**. To NIE jest prawdziwy 48-bajtowy adres. Do rozwiązania w Fazie 1.1.

### 4.3 `common/uint512/` — słowo VM

Nowy pakiet stworzony w migracji. API identyczne z `holiman/uint256`:

```go
type Int struct {
    v big.Int  // obecna implementacja — do podmiany w Fazie 1.2
}

// Operacje z semantyką mod 2^512:
func (z *Int) Add(x, y *Int) *Int
func (z *Int) Sub(x, y *Int) *Int
func (z *Int) Mul(x, y *Int) *Int
func (z *Int) Div(x, y *Int) *Int
func (z *Int) Mod(x, y *Int) *Int
func (z *Int) SDiv(x, y *Int) *Int  // signed
func (z *Int) SMod(x, y *Int) *Int
func (z *Int) Exp(base, exp *Int) *Int
func (z *Int) AddMod(x, y, m *Int) *Int
func (z *Int) MulMod(x, y, m *Int) *Int
func (z *Int) ExtendSign(x, byteNum *Int) *Int  // EVM SIGNEXTEND

// Bitowe:
func (z *Int) And(x, y *Int) *Int
func (z *Int) Or(x, y *Int) *Int
func (z *Int) Xor(x, y *Int) *Int
func (z *Int) Not(x *Int) *Int
func (z *Int) Lsh(x *Int, n uint) *Int
func (z *Int) Rsh(x *Int, n uint) *Int
func (z *Int) SRsh(x *Int, n uint) *Int  // arithmetic shift right (signed)
func (z *Int) Byte(n *Int) *Int          // zwraca n-ty bajt jako 0-255

// Porównania (unsigned):
func (z *Int) Cmp(x *Int) int    // -1, 0, 1
func (z *Int) Eq(x *Int) bool
func (z *Int) Lt(x *Int) bool
func (z *Int) Gt(x *Int) bool
func (z *Int) LtUint64(u uint64) bool
func (z *Int) GtUint64(u uint64) bool

// Porównania (signed interpretation):
func (z *Int) Slt(x *Int) bool
func (z *Int) Sgt(x *Int) bool
func (z *Int) Sign() int   // -1/0/1 interpretując jako signed 512-bit

// Stany:
func (z *Int) IsZero() bool
func (z *Int) IsUint64() bool
func (z *Int) BitLen() int

// Konwersja:
func (z *Int) SetBytes(buf []byte) *Int    // big-endian, truncate z lewej jeśli >64B
func (z *Int) SetUint64(v uint64) *Int
func (z *Int) SetFromBig(b *big.Int) bool  // returns overflow
func (z *Int) Uint64() uint64
func (z *Int) Uint64WithOverflow() (uint64, bool)
func (z *Int) Bytes() []byte     // minimalne zapisy, bez wiodących zer
func (z *Int) Bytes32() [32]byte // dolne 32 B, padding z lewej
func (z *Int) Bytes64() [64]byte // pełne 64 B, big-endian
func (z *Int) ToBig() *big.Int

// Presety:
func (z *Int) SetAllOne() *Int  // 2^512 - 1
func (z *Int) SetOne() *Int
func (z *Int) Clear() *Int
func (z *Int) Set(x *Int) *Int
func (z *Int) Clone() *Int

// Fabryki:
func NewInt(v uint64) *Int
func FromBig(b *big.Int) (*Int, bool)   // (Int, overflow)

// Fmt / debug:
func (z *Int) Hex() string
func (z Int) Format(s fmt.State, verb rune)  // fmt.Formatter
```

**Semantyka modulo:** wszystkie operacje arytmetyczne liczone mod 2^512. Wrap-around jest standardem (EVM-style).

**Signed interpretation:** używa dwu uzupełnienia. MSB (bit 511) = znak. Wartości w zakresie `[2^511, 2^512)` interpretowane jako ujemne.

**Bytes32() vs Bytes64():**
- `Bytes32()` zwraca **dolne 32 bajty** (fragment 256-256 bits). Używane tam gdzie kompatybilność z 32-bajtowymi typami (np. storage slot po ADR-001 = 32B).
- `Bytes64()` zwraca **pełne 64 bajty**. Używane dla MSTORE (zapis do pamięci 64-bajtowym słowem).

### 4.4 `core/vm/` — QRVM

To **najbardziej krytyczny moduł** projektu. Serce wykonania.

**Pliki i odpowiedzialności:**

| Plik | Rola |
|---|---|
| `qrvm.go` | Główny typ `QRVM`. Kontekst wykonania, CALL/CREATE/CREATE2. |
| `interpreter.go` | Pętla wykonania bytecode — pobranie instrukcji, wywołanie, gas. |
| `stack.go` | Stos VM (`[]uint512.Int`). |
| `memory.go` | Pamięć tymczasowa VM (`[]byte`, skalowana 64B słowami). |
| `contract.go` | Kontekst kontraktu (code, gas, caller, value). |
| `opcodes.go` | Enum opcodów, mapa nazw. |
| `jump_table.go` | Tablica operacji: dla każdego opcodu executor, koszt, stack requirements. |
| `instructions.go` | Implementacje wszystkich opcodów. |
| `instructions_test.go` | Testy opcodów. |
| `gas.go` | Obliczenia kosztu gazu (dynamic ops). |
| `gas_table.go` | Tablica kosztów, `memoryGasCost` (kwadratowa expansion). |
| `memory_table.go` | Funkcje `memorySizeFunc` per opcod (ile pamięci instrukcja potrzebuje). |
| `operations_acl.go` | Access list (EIP-2929). |
| `analysis.go` | JUMPDEST bitmap — które bajty są kodem a które danymi PUSHN. |
| `common.go` | `calcMemSize64`, `toWordSize`, `stackToAddress`. |
| `qips.go` | Opcody dodane później (SELFBALANCE, CHAINID, BASEFEE, PUSH0). |
| `contracts.go` | Precompile contracts (SHA-256, MODEXP, BN256 pairing). |
| `errors.go` | Błędy VM. |

#### 4.4.1 Stos

```go
type Stack struct {
    data []uint512.Int
}

// Operacje:
func (st *Stack) push(d *uint512.Int)
func (st *Stack) pop() uint512.Int
func (st *Stack) peek() *uint512.Int
func (st *Stack) Back(n int) *uint512.Int  // n-ty od góry
func (st *Stack) swap(n int)  // zamień top z n-tym od góry
func (st *Stack) dup(n int)   // duplikuj n-ty od góry na top
```

Limit: 1024 elementów (sprawdzane przed push przez `minStack`/`maxStack` w jump table).

Pool: `stackPool sync.Pool` — stosy są reużywane między wykonaniami żeby uniknąć alokacji.

#### 4.4.2 Pamięć

```go
type Memory struct {
    store []byte
    lastGasCost uint64
}

func (m *Memory) Set(offset, size uint64, value []byte)  // bajtowy
func (m *Memory) Set64(offset uint64, val *uint512.Int)  // całe 64B słowo (MSTORE)
func (m *Memory) Resize(size uint64)  // ekspansja
func (m *Memory) GetPtr(offset, size int64) []byte  // widok
func (m *Memory) GetCopy(offset, size int64) []byte // kopia
```

**Lazy expansion:** pamięć domyślnie ma 0 bajtów. Resize jest wywoływany w interpreterze przed wykonaniem opcodów które czytają/piszą pamięć, na podstawie ich `memorySizeFunc`.

**Koszt pamięci:** `memoryGasCost(newSize)` w `gas_table.go`:
```go
newMemSizeWords := toWordSize(newMemSize)
newMemSize = newMemSizeWords * 64  // było * 32 w EVM
// koszt = 3 * N + N^2 / 512 (stary wzór, ale z nowym N)
```

#### 4.4.3 Opcody

Plik `opcodes.go`. Każdy opcod to 1 bajt. Grupowane:

```
0x00-0x0B: arithmetic     — STOP, ADD, MUL, SUB, DIV, SDIV, MOD, SMOD, ADDMOD, MULMOD, EXP, SIGNEXTEND
0x10-0x1D: comparison/bit — LT, GT, SLT, SGT, EQ, ISZERO, AND, OR, XOR, NOT, BYTE, SHL, SHR, SAR
0x20:      KECCAK256
0x30-0x3F: closure state  — ADDRESS, BALANCE, ORIGIN, CALLER, CALLVALUE, CALLDATA*, CODE*, GASPRICE, EXTCODE*, RETURNDATA*, EXTCODEHASH
0x40-0x48: block ops      — BLOCKHASH, COINBASE, TIMESTAMP, NUMBER, PREVRANDAO, GASLIMIT, CHAINID, SELFBALANCE, BASEFEE
0x50-0x5B: storage/exec   — POP, MLOAD, MSTORE, MSTORE8, SLOAD, SSTORE, JUMP, JUMPI, PC, MSIZE, GAS, JUMPDEST
0x5F:      PUSH0
0x60-0x7F: PUSH1-PUSH32
0x80-0x9F: PUSH33-PUSH64  ← NOWE, dodane w migracji
0xA0-0xAF: DUP1-DUP16     ← przesunięte z 0x80
0xB0-0xBF: SWAP1-SWAP16   ← przesunięte z 0x90
0xC0-0xC4: LOG0-LOG4      ← przesunięte z 0xA0
0xF0-0xFF: system         — CREATE, CALL, CALLCODE, RETURN, DELEGATECALL, CREATE2, STATICCALL, REVERT, INVALID, SELFDESTRUCT
```

#### 4.4.4 Jump table

```go
type operation struct {
    execute     executionFunc
    constantGas uint64
    dynamicGas  gasFunc           // jeśli koszt zależy od runtime
    minStack    int
    maxStack    int
    memorySize  memorySizeFunc    // ile pamięci opcod potrzebuje
}

type JumpTable [256]*operation
```

Inicjalizacja w `newZondInstructionSet()`. Przykład:
```go
ADD: {
    execute:     opAdd,
    constantGas: GasFastestStep,  // = 3
    minStack:    minStack(2, 1),  // potrzebuje 2 elementy, oddaje 1
    maxStack:    maxStack(2, 1),
},
MLOAD: {
    execute:     opMload,
    constantGas: GasFastestStep,
    dynamicGas:  gasMLoad,
    minStack:    minStack(1, 1),
    maxStack:    maxStack(1, 1),
    memorySize:  memoryMLoad,  // wymaga pamięci offset + 64 bajty
},
```

Indeksowana bezpośrednio przez opcod:
```go
op := OpCode(contract.Code[pc])
operation := jumpTable[op]
```

#### 4.4.5 Interpreter — pętla główna

```go
// Uproszczony szkielet
for {
    op := contract.Code[pc]
    operation := jumpTable[op]
    
    // 1. Sprawdzenie stosu
    if stack.len() < operation.minStack { return ErrStackUnderflow }
    if stack.len() > operation.maxStack { return ErrStackOverflow }
    
    // 2. Pobranie gazu (stały + dynamiczny)
    cost := operation.constantGas
    if operation.dynamicGas != nil {
        memSize := 0
        if operation.memorySize != nil {
            memSize, _ = operation.memorySize(stack)
            memorySize = toWordSize(memSize) * 64  // było * 32
        }
        dynamicCost, err := operation.dynamicGas(qrvm, contract, stack, mem, memorySize)
        cost += dynamicCost
    }
    contract.UseGas(cost)  // zwraca ErrOutOfGas jeśli za mało
    
    // 3. Expansion pamięci
    if memorySize > 0 {
        mem.Resize(memorySize)
    }
    
    // 4. Wykonanie
    res, err := operation.execute(&pc, in, scope)
    
    // 5. Aktualizacja PC
    pc++  // domyślnie; JUMP/JUMPI modyfikują pc ręcznie
    
    if err != nil { return err }
}
```

#### 4.4.6 Instrukcje — przykłady

```go
// ADD: y = x + y (bez modyfikacji x, bo mogła być już zdjęta)
func opAdd(pc *uint64, in *QRVMInterpreter, scope *ScopeContext) ([]byte, error) {
    x, y := scope.Stack.pop(), scope.Stack.peek()
    y.Add(&x, y)
    return nil, nil
}

// MLOAD: wczytaj 64 bajty z pamięci do top stosu
func opMload(pc *uint64, in *QRVMInterpreter, scope *ScopeContext) ([]byte, error) {
    v := scope.Stack.peek()
    offset := int64(v.Uint64())
    v.SetBytes(scope.Memory.GetPtr(offset, 64))  // było 32
    return nil, nil
}

// MSTORE: zapisz 64 bajty z top stosu do pamięci
func opMstore(pc *uint64, in *QRVMInterpreter, scope *ScopeContext) ([]byte, error) {
    mStart, val := scope.Stack.pop(), scope.Stack.pop()
    scope.Memory.Set64(mStart.Uint64(), &val)  // było Set32
    return nil, nil
}

// SHL: left shift, limit = 512 (było 256)
func opSHL(pc *uint64, in *QRVMInterpreter, scope *ScopeContext) ([]byte, error) {
    shift, value := scope.Stack.pop(), scope.Stack.peek()
    if shift.LtUint64(512) {  // było 256
        value.Lsh(value, uint(shift.Uint64()))
    } else {
        value.Clear()
    }
    return nil, nil
}

// BALANCE: pobiera adres ze stosu, zastępuje balansem konta
func opBalance(pc *uint64, in *QRVMInterpreter, scope *ScopeContext) ([]byte, error) {
    slot := scope.Stack.peek()
    address := stackToAddress(slot)  // wyciąga 48B adres z 64B słowa
    slot.SetFromBig(in.qrvm.StateDB.GetBalance(address))
    return nil, nil
}
```

**`stackToAddress`** — helper do wyciągania adresu ze stosu:
```go
func stackToAddress(val *uint512.Int) common.Address {
    b := val.Bytes64()
    return common.BytesToAddress(b[16:])  // ostatnie 48 B
}
```

#### 4.4.7 Analysis (JUMPDEST bitmap)

Problem: opcod `PUSH32 0x12345...` zajmuje 33 bajty (1 opcode + 32 immediate). Te 32 bajty po PUSH32 są **danymi**, nie kodem. Jeśli PUSH32 jest poprzez `0x5B` (JUMPDEST) — to nie jest JUMPDEST!

Walidacja: pre-compute bitmap "czy ten bajt jest początkiem instrukcji". JUMP/JUMPI sprawdzają czy cel skoku to kod (bit=0) a nie dane (bit=1).

```go
// analysis.go — obliczenie bitmapy
func codeBitmapInternal(code, bits bitvec) bitvec {
    for pc := uint64(0); pc < uint64(len(code)); {
        op := OpCode(code[pc])
        pc++
        if op < PUSH1 || op > PUSH64 { continue }  // było ... op < PUSH1 || op > PUSH32
        numbits := op - PUSH1 + 1
        // ustaw bity [pc, pc+numbits) = 1 (to są dane)
        pc += uint64(numbits)
    }
    return bits
}
```

**Migracja:** rozszerzenie zakresu PUSH do PUSH64. Bufor ogona bitmapy z 4 na 8 bajtów (bo trailing PUSH64 może pisać poza faktyczny kod).

### 4.5 `core/state/` — baza stanu

**Concept:** baza wszystkich kont. Każde konto ma:
- `Nonce` uint64 — liczba wysłanych transakcji.
- `Balance` *big.Int — saldo.
- `CodeHash` Hash — hash bytecode (dla kontraktów). Pusty dla EOA.
- `Root` Hash — root storage tree (dla kontraktów).

**Trie:** Merkle-Patricia drzewo. Klucze = `Keccak256(address)` (32B), wartości = RLP konta.

```go
type Account struct {
    Nonce    uint64
    Balance  *big.Int
    Root     common.Hash  // storage root
    CodeHash []byte
}
```

**Storage per kontrakt:** też Merkle-Patricia, klucze = `Keccak256(slot_index)`, wartości = RLP(slot_value). `Account.Root` to root tego drzewa.

#### 4.5.1 StateDB

```go
type StateDB struct {
    db          Database        // underlying trie DB
    trie        Trie            // state trie
    stateObjects map[common.Address]*stateObject  // in-memory cache
    journal     *journal        // rewertowalne zmiany
    // ...
}

// Operacje:
func (s *StateDB) GetBalance(addr common.Address) *big.Int
func (s *StateDB) SetBalance(addr common.Address, amount *big.Int)
func (s *StateDB) AddBalance(addr common.Address, amount *big.Int)
func (s *StateDB) GetNonce(addr common.Address) uint64
func (s *StateDB) SetNonce(addr common.Address, nonce uint64)
func (s *StateDB) GetCode(addr common.Address) []byte
func (s *StateDB) SetCode(addr common.Address, code []byte)
func (s *StateDB) GetState(addr common.Address, key common.Hash) common.Hash
func (s *StateDB) SetState(addr common.Address, key, val common.Hash)
// ...
```

**Journal** — każda zmiana stanu jest zapisywana do listy "co cofnąć", żeby można było wrócić do snapshotu po REVERT:

```go
snapshot := state.Snapshot()
// ... modyfikacje
if bad { state.RevertToSnapshot(snapshot) }
```

#### 4.5.2 stateObject

Pojedyncze konto w pamięci:
```go
type stateObject struct {
    address  common.Address
    addrHash common.Hash        // = Keccak256(address[:])
    data     Account            // nonce, balance, root, codeHash
    code     []byte
    originStorage  Storage      // what was in DB
    dirtyStorage   Storage      // pending writes
    // ...
}
```

**Obserwacja migracyjna:** `addrHash = Keccak256(address[:])` → z 48-bajtowego adresu liczymy 32-bajtowy hash. **Działa automatycznie** bo Keccak256 akceptuje dowolne wejście.

### 4.6 `core/types/` — bloki, transakcje, receipty, logi

**Pliki:**
- `block.go` — `Block`, `Header`.
- `transaction.go` — `Transaction`, subtypy (legacy, EIP-2930, EIP-1559).
- `tx_dynamic_fee.go` — EIP-1559 tx.
- `receipt.go` — `Receipt` (wynik wykonania).
- `log.go` — `Log` (event).
- `transaction_signing.go` — `Signer` (EIP-155, EIP-2930).

**Struktura Block:**

```go
type Header struct {
    ParentHash  common.Hash
    Coinbase    common.Address   // 48B automatycznie
    Root        common.Hash      // state root po bloku
    TxHash      common.Hash      // Merkle root transakcji
    ReceiptHash common.Hash      // Merkle root receiptów
    Bloom       Bloom            // bloom filter logów
    Difficulty  *big.Int
    Number      *big.Int
    GasLimit    uint64
    GasUsed     uint64
    Time        uint64
    Extra       []byte
    MixDigest   common.Hash
    Nonce       BlockNonce
    BaseFee     *big.Int         // EIP-1559
}

type Block struct {
    header       *Header
    transactions Transactions
    // ...
}
```

**RLP encoding:** wszystkie typy mają `EncodeRLP` / `DecodeRLP`. `common.Address` jest RLP-encodowany jako string 48B (było 20B) — **automatycznie**.

**Hash bloku:** `hash = Keccak256(RLP(header))`. Różny niż w EVM, bo header zawiera 48B Coinbase + Extra zmienione.

**Transaction:**

```go
type Transaction struct {
    inner TxData  // interfejs: LegacyTx, AccessListTx, DynamicFeeTx
    // ...
}

type DynamicFeeTx struct {
    ChainID    *big.Int
    Nonce      uint64
    GasTipCap  *big.Int
    GasFeeCap  *big.Int
    Gas        uint64
    To         *common.Address  // nil = contract creation
    Value      *big.Int
    Data       []byte
    AccessList AccessList
    V, R, S    *big.Int  // signature
}
```

**Signer:**

```go
type Signer interface {
    Sender(tx *Transaction) (common.Address, error)
    Hash(tx *Transaction) common.Hash
    SignatureValues(tx *Transaction, sig []byte) (r, s, v *big.Int, err error)
    ChainID() *big.Int
}
```

**Flow podpisu:**
1. `Hash(tx)` — policz hash "zasypki" (tx fields bez sig).
2. Sign(hash) → (r, s, v) — ECDSA (legacy) albo ML-DSA-87 (produkcyjne).
3. Zapisz (r, s, v) w tx.

**Flow weryfikacji:**
1. `Hash(tx)` — ten sam hash co przy podpisie.
2. Recover pubkey z (r, s, v, hash) → pubkey.
3. `PubkeyToAddress(pubkey)` → adres wysyłającego.

**Log:**

```go
type Log struct {
    Address     common.Address  // kontrakt który emituje (48B automatycznie)
    Topics      []common.Hash   // 0..4 topików 32B (ADR-002)
    Data        []byte          // arbitrary payload
    BlockNumber uint64
    TxHash      common.Hash
    TxIndex     uint
    BlockHash   common.Hash
    Index       uint
    Removed     bool            // true jeśli usunięty przez reorg
}
```

**Receipt:**

```go
type Receipt struct {
    Type              uint8
    PostState         []byte
    Status            uint64
    CumulativeGasUsed uint64
    Bloom             Bloom
    Logs              []*Log
    TxHash            common.Hash
    ContractAddress   common.Address  // 48B (dla CREATE/CREATE2)
    GasUsed           uint64
    EffectiveGasPrice *big.Int
    BlockHash         common.Hash
    BlockNumber       *big.Int
    TransactionIndex  uint
}
```

### 4.7 `accounts/abi/` — ABI

**Co to ABI:** standard kodowania argumentów i wartości zwracanych funkcji kontraktu. Między Go (wywołującym kontrakt z zewn.) a EVM (wykonującym kontrakt) potrzebna jest wspólna reprezentacja binarna parametrów.

**Format EVM:** "32-bajtowe sloty" (w naszym VM **64-bajtowe**). Każdy parametr zajmuje co najmniej jeden slot; dłuższe (bytes, string, arrays) są kodowane head+tail.

**Przykład:**

Sygnatura:
```solidity
function transfer(address to, uint256 amount) returns (bool)
```

Wywołanie `transfer(0xABC..., 100)`:

```
Calldata:
    selector(4B):           keccak256("transfer(address,uint256)")[:4] = 0xa9059cbb
    arg0 (address): 64B:    zeros(16) || addr_48B        (było: zeros(12) || addr_20B)
    arg1 (uint256): 64B:    zeros(32) || uint256_32B     (było: uint256_32B)
```

Wartość zwrotu (returndata):
```
Returndata:
    bool: 64B:              zeros(63) || 0x01  lub  zeros(63) || 0x00
```

**Pliki:**
- `type.go` — typy ABI (AddressTy, UintTy, IntTy, BoolTy, BytesTy, StringTy, ArrayTy, SliceTy, TupleTy, FixedBytesTy, FunctionTy, HashTy).
- `argument.go` — `Arguments` — kolekcja parametrów, metody Pack/Unpack.
- `pack.go` — implementacja Pack (Go wartość → bajty).
- `unpack.go` — implementacja Unpack (bajty → Go wartość).
- `topics.go` — konstruowanie/parsowanie topiców eventów.
- `event.go`, `method.go`, `error.go` — reprezentacja events/methods/errors.
- `abi.go` — główny typ `ABI`, `Pack(methodName, args...)`.

**Rozmiar slotu po migracji:**

```go
// type.go:
func getTypeSize(t Type) int {
    if t.T == ArrayTy && !isDynamicType(*t.Elem) {
        return t.Size * 64  // było 32
    }
    // ...
    return 64  // było 32
}

// pack.go:
func packNum(value reflect.Value) []byte {
    return math.U512Bytes(big.NewInt(value.Int()))  // było U256Bytes
}
```

**Address padding:** address 48B w 64B slocie = 16 zer z lewej + 48 bajtów adresu.

**FunctionTy:** adres 48B + selector 4B = 52B. Było 20+4 = 24B. Zmiana w `readFunctionType`:
```go
func readFunctionType(t Type, word []byte) (funcTy [common.AddressLength + 4]byte, err error) {
    // ... waliduje że bajty po 52. są zera
    copy(funcTy[:], word[0:common.AddressLength+4])
}
```

### 4.8 `internal/qrlapi/` — JSON-RPC

Implementacja API zgodnego z Ethereum JSON-RPC (`qrl_*` namespace).

**Pliki:**
- `api.go` — główne metody: `qrl_blockNumber`, `qrl_getBalance`, `qrl_sendRawTransaction`, etc.
- `transaction_args.go` — typ `TransactionArgs` — parametry w `qrl_call`, `qrl_estimateGas`.
- `backend.go` — interfejs do backendu (łańcuch, state, txpool).

**Serializacja adresu w JSON:**

```go
// Address ma MarshalText/UnmarshalText używające hex
func (a Address) MarshalText() ([]byte, error) {
    return hexutil.BytesQ(a[:]).MarshalText()
}
```

Zwraca `"Q" + 96 hex`. JSON reprezentacja automatycznie skaluje się z `AddressLength`.

**Przykład odpowiedzi:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "blockHash": "0xabc...",
    "blockNumber": "0x1234",
    "from": "Q000000000000000000000000000000000000000000000000000000005aaeb6053f3e94c9b9a09f33669435e7ef1beaed",
    "gas": "0x5208",
    "gasPrice": "0x...",
    "hash": "0xdef...",
    "to": "Q...",
    "value": "0x...",
    ...
  }
}
```

---

## 5. Cross-cutting: RLP, Keccak, Gas

### 5.1 RLP (Recursive Length Prefix)

Binarny format serializacji używany wszędzie w Ethereum/QRL. Kluczowy bo: **RLP-encoded bajt = kanoniczna reprezentacja = hashowalny input**. Nawet minimalna zmiana (np. 48B vs 20B address) zmienia RLP → zmienia hash → zmienia state root → rozjazd między węzłami.

**Zasady:**
- Bajt `0x00..0x7f` = sam siebie.
- String 0..55 bajtów: `0x80 + len, data...`
- String >55 bajtów: `0xb7 + len(len), len, data...`
- Lista 0..55 bajtów: `0xc0 + len, items...`
- Lista >55 bajtów: `0xf7 + len(len), len, items...`

**Automatyczne dla Address:**
```go
// RLP encoder traktuje [48]byte jako 48-bajtowy string
// Wynik dla address = 0xABC...:
// 0x94 + 48 bajtów danych (było 0x94 + 20 bajtów → teraz 0xB0 + 48 bajtów)
// (0x94 = 0x80 + 0x14 = 20-byte string; 0xB0 = 0x80 + 0x30 = 48-byte string)
```

### 5.2 Keccak-256 vs Keccak-512 vs SHA-3

**Keccak** — rodzina funkcji hashujących bazująca na konstrukcji "sponge". Pre-FIPS-202.

Ethereum używa **legacy Keccak-256** — Keccak z output 256 bitów, pre-FIPS padding. To **NIE jest SHA3-256** (który ma inny padding wg FIPS-202) — różnią się jednym bitem, co daje różne wartości hasha.

W go-qrl:
- `crypto.Keccak256` (plik `crypto/sha3.go`) — używa `golang.org/x/crypto/sha3.NewLegacyKeccak256()`.
- Nasze `crypto.Keccak256Hash` — wrapuje w `common.Hash`.

Po migracji Keccak-512 może być używany przy derivacji adresów kontraktów, ale nie przy formatowaniu adresu tekstowego.

### 5.3 Gas metering

Gas = koszt wykonania. Każda instrukcja ma zdefiniowany koszt. Transakcja ma `gasLimit` — jeśli wykonanie przekracza, wszystko jest rollback (oprócz zapłaconego gazu).

**Typy kosztu:**
- **Stały per instrukcja** — `constantGas` w jump table. Np. ADD = 3, MUL = 5, KECCAK256 = 30.
- **Dynamiczny** — `dynamicGas` — zależy od runtime. Np. KECCAK256 dodatkowo `6 * wordsOfInput`, EXP dodatkowo `50 * bytesOfExponent`.
- **Memory expansion** — kwadratowy koszt rosnącej pamięci.
- **Storage** — SSTORE/SLOAD, cold/warm access, refunds.
- **Call** — 63/64 reguła, stipend dla value transfer.

**Stałe w `params/gas.go`:**
```go
const (
    GasQuickStep      uint64 = 2   // POP, PC, MSIZE
    GasFastestStep    uint64 = 3   // ADD, SUB, AND, OR, XOR, NOT, LT, GT, ...
    GasFastStep       uint64 = 5   // MUL, DIV, SDIV, MOD, SMOD, SIGNEXTEND, ...
    GasMidStep        uint64 = 8   // ADDMOD, MULMOD, JUMP
    GasSlowStep       uint64 = 10  // JUMPI
    GasExtStep        uint64 = 20  // BLOCKHASH
    // ...
    SstoreSetGas      uint64 = 20000
    SstoreResetGas    uint64 = 5000
    SstoreClearGas    uint64 = 5000
    // ...
)
```

**Po migracji:** stałe zachowane. Zmiany w `toWordSize * 32 → * 64` i `newMemSize * 32 → * 64`. Pełna rekalibracja w ADR-005.

---

## 6. Relacja do qrvmone / qrvmc

### qrvmc (EVMC dla QRL)

Nagłówek C ABI do wywoływania VM z hosta. Analog `evmc.h` z Ethereum.

Kluczowe typy:
```c
typedef struct qrvmc_address { uint8_t bytes[48]; } qrvmc_address;
typedef struct qrvmc_bytes32 { uint8_t bytes[32]; } qrvmc_bytes32;
```

- `qrvmc_address` = 48 bajtów ✓
- `qrvmc_bytes32` = 32 bajty (klucze/wartości storage, topics, CREATE2 salt)

Struktura wywołania:
```c
qrvmc_result qrvmc_execute(
    const qrvmc_instance*,
    const qrvmc_host_interface*,
    qrvmc_host_context*,
    qrvmc_revision,
    const qrvmc_message*,  // kind, recipient, sender, value, input, gas, ...
    const uint8_t* code,
    size_t code_size
);
```

Host implementuje `qrvmc_host_interface` (callbacks dla storage, balance, create, call).

### qrvmone (evmone dla QRL)

Wydajna implementacja VM w C++. Wywoływana przez hosta przez qrvmc.

Kluczowe pliki które mapują na nasze:
- `lib/qrvmone/execution_state.hpp` — `uint512` stack (odpowiednik `core/vm/stack.go`).
- `lib/qrvmone/instructions.hpp` — implementacje opcodów (odpowiednik `core/vm/instructions.go`).
- `lib/qrvmone/instructions_opcodes.hpp` — enum opcodów (odpowiednik `core/vm/opcodes.go`).
- `lib/qrvmone/baseline.cpp` — code padding, JUMPDEST analysis (odpowiednik `core/vm/analysis.go`).
- `lib/qrvmone/instructions_storage.cpp` — SSTORE/SLOAD.
- `lib/qrvmone/instructions_calls.cpp` — CALL/CREATE/CREATE2.

**Stan migracji qrvmone:**
- Stack `uint512`: ✓
- Opcody PUSH33-64, DUP/SWAP/LOG 0xA0/0xB0/0xC0: ✓
- Code padding 64+1: ✓
- Storage: 32B slot (zgodne z rekomendacją ADR-001)
- Topics: 32B (zgodne z rekomendacją ADR-002)
- CREATE2 salt: 32B (zgodne z rekomendacją ADR-003)

### Dlaczego dwie implementacje?

- **go-qrl** (Go): referencyjna, łatwa w utrzymaniu, debugowaniu.
- **qrvmone** (C++): wydajna, używana w produkcyjnych klientach dla high-throughput.

**Kluczowe:** obie muszą dawać **identyczne rezultaty**. Rozbieżność = consensus fail. Dlatego Faza 1.3 (cross-validation) i Faza 5.1 (differential fuzzing) są krytyczne.

---

## 7. Struktura testów

### 7.1 Typy testów w go-qrl

| Typ | Lokalizacja | Przykład | Zakres |
|---|---|---|---|
| Unit | `*_test.go` obok kodu | `crypto/crypto_test.go:TestNewContractAddress` | 1 funkcja |
| Integration | często `core/*_test.go` | `core/blockchain_test.go:TestLogReorgs` | wiele modułów |
| Fixture-based | `testdata/*.json` | `core/vm/testdata/testcases_add.json` | data-driven |
| Fuzzer | `*_fuzz_test.go` | `core/state/statedb_fuzz_test.go` | random input |
| Benchmark | `*_test.go` z `Benchmark*` | `crypto/crypto_test.go:BenchmarkKeccak256Hash` | performance |

### 7.2 Problem fixtur w migracji

Wiele testów miało hardcoded 20-bajtowe adresy, 32-bajtowe sloty, RLP bloki z adresami. Po migracji te fixtury są **niepoprawne** dla nowego kodu (ale kod działa poprawnie — fixture mówi "test expects X" gdzie X bazuje na starych wartościach).

**Obecne rozwiązanie (tymczasowe):** `t.Skip(...)` w testach z TODO. ~54 testów w kodzie który migrowaliśmy, plus ~150-250 w pakietach nie dotkniętych.

**Docelowe rozwiązanie (Faza 3):** generatory w `hack/gen/*.go` regenerują wszystkie fixtury deterministycznie z bieżącego kodu. Każdy test ma ścieżkę "jak zregenerować".

### 7.3 Jakie testy pozostają zielone bez zmian

- Testy które **nie zakładają konkretnej wartości adresu/hashe** a sprawdzają **zachowanie**. Np. "po A+B wartość na stosie to C" — jeśli A, B są liczbowe i wynik też, żadna zmiana.
- Testy używające typów strukturalnych bez konkretnych hex literałów.

### 7.4 Jakie trzeba regenerować

- Testy z hex literałami adresów/hashów.
- Testy porównujące do hardcoded RLP bajtów.
- Testy gas-usage z konkretnymi wartościami (bo gas rekalibracja).
- Testy kanonikalizacji adresów `Q` + 96 lowercase hex.

---

## 8. Struktura projektu na wysokim poziomie

```
go-qrl/
├── cmd/                 # programy CLI
│   ├── gqrl/            # główny klient (node)
│   ├── clef/            # standalone signer
│   ├── qrvm/            # debugger bytecode
│   ├── qrlkey/          # narzędzie do kluczy
│   └── ...
├── accounts/
│   ├── abi/             # ABI encoding
│   └── keystore/        # keystore plików
├── common/
│   ├── types.go         # Address, Hash
│   ├── big.go           # Big0, Big1, ...
│   ├── uint512/         # NOWE — 512-bit int dla VM
│   └── math/            # PaddedBigBytes, U256Bytes, U512Bytes (nowe)
├── consensus/
│   └── beacon/          # Proof-of-Stake
├── core/
│   ├── vm/              # QRVM — stack, memory, instructions, interpreter
│   ├── state/           # StateDB, stateObject
│   ├── types/           # Block, Transaction, Receipt, Log
│   ├── txpool/          # tx pool
│   ├── rawdb/           # low-level DB access
│   ├── blockchain.go    # zarządzanie łańcuchem
│   ├── genesis.go       # genesis block
│   └── state_processor.go  # apply(block, state)
├── crypto/
│   ├── crypto.go        # Keccak, CreateAddress, CreateAddress2, PubkeyToAddress
│   ├── secp256k1/       # ECDSA (legacy)
│   └── pqcrypto/        # ML-DSA-87 post-quantum
├── internal/
│   └── qrlapi/          # JSON-RPC handlers
├── params/
│   ├── config.go        # ChainConfig, ChainID, genesis hashes
│   └── gas.go           # stałe gazu
├── qrl/                 # QRL protocol
│   ├── backend.go
│   ├── downloader/      # sync
│   ├── filters/         # log filters
│   ├── protocols/       # p2p protocols
│   └── tracers/         # execution tracers
├── p2p/                 # P2P transport
├── qrldb/               # DB interfejs
├── rlp/                 # RLP encoder/decoder
├── signer/              # clef signer
├── trie/                # Merkle-Patricia trie
└── docs/
    ├── 48-BYTE-ADDRESS-MIGRATION.md  # oryginalny plan (historyczny)
    ├── PRODUCTION-PLAN.md            # obecny plan
    ├── ADR-ANALYSIS.md               # analiza decyzji
    └── ARCHITECTURE-GO-QRL.md        # ten dokument
```

### Grubość modułów (przykładowe liczby)

| Moduł | LoC (est.) | Testy (est.) |
|---|---|---|
| `core/vm` | ~5000 | ~3000 |
| `core/state` | ~4000 | ~4500 (fuzz included) |
| `core/types` | ~3500 | ~2500 |
| `accounts/abi` | ~2500 | ~3500 |
| `crypto` | ~2000 | ~800 |
| `common` | ~1500 | ~500 |
| `internal/qrlapi` | ~4000 | ~3300 |

Większość kodu dotkniętego bezpośrednio migracją: ~**20k LoC produkcyjnego + ~15k testów**.

---

## 9. Gdzie zacząć czytać kod

Dla zrozumienia **wykonania transakcji** (najbardziej krytyczna ścieżka):

1. `core/state_processor.go:Process` — apply block to state.
2. `core/state_transition.go:TransitionDb` — apply single tx.
3. `core/vm/qrvm.go:QRVM.Call` / `.Create` — wejście do VM.
4. `core/vm/interpreter.go:QRVMInterpreter.Run` — pętla interpretera.
5. `core/vm/instructions.go` — implementacje opcodów.

Dla zrozumienia **storage**:

1. `core/state/statedb.go:StateDB.GetState/SetState`.
2. `core/state/state_object.go:stateObject.setState`.
3. `trie/trie.go:Trie.Update`.

Dla zrozumienia **ABI**:

1. `accounts/abi/abi.go:ABI.Pack` (wysłanie do kontraktu).
2. `accounts/abi/pack.go:packElement`.
3. `accounts/abi/abi.go:ABI.Unpack` (dekodowanie wyjścia).
4. `accounts/abi/unpack.go:toGoType`.

Dla zrozumienia **genesis + chain setup**:

1. `core/genesis.go:DefaultGenesisBlock` / `DefaultTestnetGenesisBlock`.
2. `params/config.go` — ChainID, genesis hashes.
3. `cmd/gqrl/main.go` — startup klienta.

---

## 10. Debugging tips

### Tracing wykonania VM

```go
// core/vm/logger.go (jeśli uruchomiasz z flag --trace)
type StructLog struct {
    Pc            uint64
    Op            OpCode
    Gas           uint64
    GasCost       uint64
    Memory        []byte
    MemorySize    int
    Stack         []uint512.Int  // było uint256
    ReturnData    []byte
    Storage       map[Hash]Hash
    Depth         int
    RefundCounter uint64
    Err           error
}
```

Trace pokaże dokładnie co dzieje się w każdym opcodzie.

### Inspekcja stanu przez RPC

```bash
# Balance konta
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"qrl_getBalance","params":["Q...","latest"],"id":1}' \
  http://localhost:8545

# Storage slot
curl ... "qrl_getStorageAt" ["Q...", "0x0", "latest"]

# Pobranie kodu kontraktu
curl ... "qrl_getCode" ["Q...", "latest"]
```

### Debugowanie fixtur

Jeśli test failuje z "expected X got Y", często można policzyć Y z bieżącego kodu i wstawić:

```go
// Dodaj do testu:
t.Logf("actual: %x", actualValue)
// Uruchom, skopiuj output, wklej jako expected
```

Tak robiliśmy w Fazie 5 dla `TestDump`/`TestIterativeDump`.

---

## 11. Glosariusz (dla osób z Ethereum background)

| EVM/Ethereum | QRL/go-qrl | Różnica |
|---|---|---|
| EVM | QRVM | 512-bit słowo, 48B adresy |
| geth | gqrl | rebrand |
| ether | QRL token (nazwa ustalana) | token natywny |
| wei | planck | najmniejsza jednostka (QRL ma 10^19 = quintillion, "planck" to QRL-term) |
| secp256k1 | ML-DSA-87 | post-quantum, większe klucze i podpisy |
| evmc | qrvmc | — |
| evmone | qrvmone | — |
| address (20B) | address (48B) | 2.4× większy |
| stack word (256b) | stack word (512b) | 2× większy |
| PUSH1-PUSH32 | PUSH1-PUSH64 | extended |
| Chain ID 1 (mainnet) | TBD | nowy |
