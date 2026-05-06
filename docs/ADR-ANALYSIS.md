# Analiza decyzji architektonicznych (ADR) — migracja 48B adresów + 512-bit VM

> Dokument uzupełniający do `PRODUCTION-PLAN.md`. Dla każdego ADR szczegółowo opisuje problem, tło teoretyczne, opcje i rekomendacje. Skierowany do programistów, którzy nie znają specyfiki blockchainów / EVM / go-qrl, ale są zaawansowanymi inżynierami.

## Jak używać tego dokumentu

- Czytaj liniowo — kolejne ADR-y budują na pojęciach wprowadzonych wcześniej.
- Każdy ADR ma sekcje: **Tło**, **Problem**, **Obecny stan**, **Opcje**, **Analiza**, **Rekomendacja**.
- Przykłady kodu są ilustracyjne (czasem pseudo-Solidity).
- Decyzje zapisujemy w [`PRODUCTION-PLAN.md`](./PRODUCTION-PLAN.md) — ten dokument to analiza, nie rejestr.

---

## Słownik podstawowy

- **EVM** (Ethereum Virtual Machine) — stack-based maszyna wirtualna wykonująca bytecode kontraktów. Operacje na 256-bitowych słowach. QRL ma **QRVM** — widełkę EVM.
- **Kontrakt** — deterministyczny program uruchamiany przez VM, identyfikowany przez adres.
- **Adres** — 20B (EVM) / 48B (QRL) identyfikator konta lub kontraktu.
- **Storage** — persystentna pamięć per-kontrakt: `mapa(klucz 32B → wartość 32B)`.
- **State** — zbiorcza baza wszystkich kont: `mapa(adres → stan konta)`, gdzie stan = nonce + balance + codeHash + storageRoot.
- **Trie** — Merkle-Patricia trie, struktura danych używana do przechowywania state z kryptograficznym podsumowaniem (root hash).
- **Stos VM** — do 1024 elementów, każdy 256-bit (EVM) / 512-bit (nasz QRVM). Wszystkie operacje arytmetyczne/logiczne działają na wierzchu stosu.
- **Pamięć VM** — tymczasowa pamięć bajtowa per-wykonanie, skalowana 32-bajtowymi słowami (EVM) / 64-bajtowymi (nasz QRVM).
- **Calldata** — bajty wejściowe przekazane do funkcji kontraktu przy wywołaniu.
- **ABI** (Application Binary Interface) — standard kodowania argumentów i wartości zwracanych funkcji kontraktu.
- **Gas** — jednostka kosztu wykonania. Każda instrukcja kosztuje gaz, a transakcja ma limit. Chroni przed nieskończonymi pętlami i DoS.
- **Opcode** — 1-bajtowa instrukcja w bytecode VM (np. 0x01 = ADD, 0x52 = MSTORE).
- **Transakcja** — podpisana wiadomość wywołująca zmianę stanu (transfer, deploy, call).
- **Blok** — zbiór transakcji plus metadane, linkowane hash-pointerami w łańcuch.
- **Receipt** — wynik wykonania transakcji (status, gas used, logi).
- **Log / Event** — zdarzenie emitowane przez kontrakt, indeksowane w drzewie logów.
- **Topic** — do 4 × 32B tagów logu, używane do filtrowania.
- **RLP** (Recursive Length Prefix) — binarny format serializacji używany w Ethereum/QRL dla bloków, tx, trie nodes.
- **Keccak-256** — kryptograficzny hash 256-bit używany wszędzie w EVM. Keccak-512 to jego 512-bit wariant.

---

## ADR-001: Storage slot width (szerokość slotu pamięci trwałej)

### Tło

Każdy smart-kontrakt na maszynie wirtualnej ma własną persystentną pamięć, niezależną od innych kontraktów. Nazywamy ją **storage**. Struktura to płaska mapa klucz-wartość:

```
storage: mapa (klucz → wartość)
```

Jednostka tej mapy nazywa się **slot**. Slot to para:
- **slot_index** (klucz) — adres w obrębie storage kontraktu
- **slot_value** (wartość) — dana zapisana pod tym adresem

W klasycznym EVM:
- slot_index = 256 bitów (32 bajty)
- slot_value = 256 bitów (32 bajty)
- kontrakt ma 2^256 slotów (teoretycznie), domyślnie puste (= 0)

### Opcody storage

EVM ma dwa opcody dla storage:
- **SSTORE** (0x55): zdejmij z stosu dwie wartości — klucz i wartość — zapisz `storage[klucz] = wartość`.
- **SLOAD** (0x54): zdejmij z stosu klucz, zapisz na wierzch stosu `storage[klucz]`.

### Przykład działania

Solidity:
```solidity
contract Foo {
    uint256 a;        // zmienna przypisana do slotu 0
    uint256 b;        // zmienna przypisana do slotu 1
    uint256[3] arr;   // zmienne w slotach 2, 3, 4
    mapping(address => uint256) balances;  // mapowanie od slotu 5
}
```

Kompilator Solidity automatycznie przypisuje zmienne do slotów. Dla mapowań (dynamiczny rozmiar) klucz slotu liczy się jako hash:

```
slot_dla_balances[addr] = keccak256(addr || 5)   // 5 to "baza" slotu mapowania
```

Wykonanie `a = 42` generuje:
```
PUSH1 42       ; push value 42
PUSH1 0        ; push slot index 0
SSTORE         ; storage[0] = 42
```

Wykonanie `uint x = a` generuje:
```
PUSH1 0        ; push slot index 0
SLOAD          ; push storage[0]
```

### Wizualizacja

```
Storage kontraktu X:
┌──────────────────────────────┬──────────────────────────────┐
│ slot_index (32 B)            │ slot_value (32 B)            │
├──────────────────────────────┼──────────────────────────────┤
│ 0x0000...00 (=0)             │ 0x0000...2a (=42)            │  ← a
│ 0x0000...01 (=1)             │ 0x0000...00 (=0)             │  ← b (nieset)
│ 0x0000...02 (=2)             │ 0x0000...07 (=7)             │  ← arr[0]
│ 0xabc1...ef (hash)           │ 0x0000...64 (=100)           │  ← balances[0xAA...]
└──────────────────────────────┴──────────────────────────────┘
```

### Problem w naszej migracji

Stos VM został poszerzony z 256-bit na 512-bit (64 B). Ale state DB wciąż operuje na 32-bajtowych `common.Hash` (klucz) i 32-bajtowej wartości.

Obecna implementacja `opSstore`:
```go
loc := stack.pop()    // uint512 (64 B) ze stosu
val := stack.pop()    // uint512 (64 B) ze stosu
scope.Contract.StateDB.SetState(addr, loc.Bytes32(), val.Bytes32())
//                                          ↑               ↑
//                               bierze DOLNE 32 bajty, górne ODRZUCONE
```

### Opcje

#### Opcja A — slot pozostaje 32 B (obecna implementacja, zgodna z qrvmone)

```go
// SSTORE obcina dolne 32 B, górne 32 B są tracone
SetState(addr, val.Bytes32(), ...)
```

**Konsekwencje semantyczne:**

1. **Utrata górnych bitów przy zapisie:**
   ```solidity
   uint512 x = 2^400;     // wartość > 2^256
   storage[slot] = x;     // obcięte do 0 (górne bity)
   uint512 y = storage[slot];   // odczyt daje 0, nie 2^400
   ```

2. **Kolizja kluczy dla mapowań uint512:**
   ```solidity
   mapping(uint512 => uint) m;
   m[2^300] = 1;           // klucz ma te same dolne 256 bitów co...
   m[2^300 + 2^256] = 2;   // ...ten klucz. Nadpisanie!
   ```

#### Opcja B — slot poszerzony do 64 B

- state DB używa 64-bajtowych kluczy i wartości
- `SetState(addr, [64]byte, [64]byte)`
- trie potrzebuje innego layout (szerszy klucz, ewentualnie Keccak-512 dla bezpieczeństwa)

### Analiza trade-offów

| Aspekt | 32 B (Opcja A) | 64 B (Opcja B) |
|---|---|---|
| Zgodność z EVM/Solidity | wysoka (Solidity używa 32B slotu natywnie) | niska (kompilator wymaga adaptacji) |
| Zgodność z qrvmone | ✅ identyczna | wymaga zmian w C++ |
| Trie schema | bez zmian | wymaga projektu |
| Gas SSTORE | bez zmian | prawdopodobnie ×2 |
| Hash klucza trie | Keccak256 (32B) wystarczy | Keccak512 (64B) być może |
| Faithfulness 512-bit | brak (obcięcie) | pełna |
| Ryzyko bugów | ciche obcięcia | koszty gazu niespodzianki |
| Narzędzia zewn. (exploratory, debugery) | działają jak ethereumowe | wymagają update |

### Ukryte założenie: co znaczy "512-bit VM"?

Tu jest sedno: **czy VM jest "512-bit" tylko obliczeniowo, czy też persystentnie?**

- Obliczeniowo: ALU (stack), tymczasowa pamięć (memory), wartości w registrach → 512-bit. Bo chcemy post-quantum kryptografii, większych hashów, większych adresów. To nie musi wpływać na storage.
- Persystentnie: storage slot, trie klucze → 512-bit. To oznacza 2× większe zapisy na dysk, inną semantykę.

Nie ma obowiązku żeby te dwie szerokości były równe. W klasycznym EVM są (obie 256-bit), ale to przypadek historyczny, nie wymóg logiczny.

### Rekomendacja: **Opcja A (32 B)**

**Argumenty:**

1. **Zgodność z qrvmone.** qrvmone już wybrał 32B (`instructions_storage.cpp`), więc tę decyzję trzeba by retro-zmieniać w C++ — dodatkowa praca, potencjalne niespójności w trakcie migracji.

2. **Storage ma inny cel niż stos.** Stos to arytmetyka — potrzebuje pełnej precyzji obliczeniowej. Storage to pamięć trwała — kluczem jest trwałość i efektywność, nie szerokość słowa. Nie ma dowodu że 512-bitowa wartość w storage jest realnie potrzebna (kryptograficzne primatywy, które chcemy przechowywać — np. hashe — są w zakresie 256-bit).

3. **Gas.** Przechodzenie na 64B storage ~2× droższy SSTORE, co zmienia ekonomię kontraktów nieproporcjonalnie do zysku.

4. **Możliwość ewolucji.** Jeśli kiedyś 512-bitowy storage będzie potrzebny, można **dodać** nowe opcody `SSTORE64` / `SLOAD64` obok istniejących, bez łamania kompatybilności. Trudniej ten kierunek odwrócić.

5. **Nawet w przyszłości Solidity-like** — kompilator prawdopodobnie będzie emitował dwa zapisy 32B dla wartości > 256-bit, bo tak działają kolory danych w języku wysokopoziomowym. 64B slot nie daje tu oczywistej wartości.

**Kontrargumenty (jeśli ktoś silnie forsuje B):**
- Symetria stack/memory/storage jest elegancka.
- Chroni programistę przed cichym obcinaniem.
- Upraszcza formalną weryfikację kontraktu.

Jeśli te argumenty przeważą — OK, ale trzeba zaplanować ~2 tygodnie pracy nad C++ i storage trie + rekalibrację gazu.

---

## ADR-002: Log topic width (szerokość topiku logu)

### Tło

Kontrakt może **emitować zdarzenia** (events, w EVM: "logs"). Zdarzenia trafiają do receipt transakcji, indeksowane, i służą frontom aplikacji do nasłuchiwania zmian stanu bez ciągłego polowania.

Log ma strukturę:
```
log = {
    address:  adres kontraktu, który emitował (48 B)
    topics:   tablica 0..4 elementów, każdy to tag (32 B)
    data:     binarny payload (dowolna długość)
}
```

Opcody: `LOG0`, `LOG1`, `LOG2`, `LOG3`, `LOG4` — różnią się liczbą topików. Działanie `LOG_N`:

1. Zdejmij z stosu offset i size (gdzie w pamięci jest `data`).
2. Zdejmij N topików (każdy pobierany jako 32B ze stosu).
3. Dodaj log do receipta.

### Po co to topicy?

Topicy są **indeksowalne** — można po nich szybko wyszukiwać. Klasyczny wzorzec:
- topic[0] = hash sygnatury zdarzenia (np. `keccak256("Transfer(address,address,uint256)")`)
- topic[1] = `from` address (hashowany lub padded)
- topic[2] = `to` address
- topic[3] = dodatkowy identyfikator

Kod Solidity:
```solidity
event Transfer(address indexed from, address indexed to, uint256 value);

function transfer(address to, uint256 amount) external {
    emit Transfer(msg.sender, to, amount);
}
```

Kompilator zamienia to na `LOG3` z:
- topic[0] = `keccak256("Transfer(address,address,uint256)")`
- topic[1] = `msg.sender` (32B padded z adresu)
- topic[2] = `to` (32B padded)
- data = abi.encode(amount)

Frontend może potem szybko filtrować "wszystkie Transfer z topic[1] = 0xABC..." bez skanowania całej historii.

### Problem w naszej migracji

Adresy mają teraz 48 bajtów. Topic ma 32 bajty. Adres **nie mieści się** w topiku.

Obecne obejście w `accounts/abi/topics.go`:
```go
case common.Address:
    // Address (48 bytes) is larger than Hash (32 bytes), so we hash it for the topic.
    copy(topic[:], crypto.Keccak256(rule[:]))
```

Czyli adres jest **hashowany** do 32 bajtów i to hashes leci do topika. Skutek:
- Filtrowanie "znajdź wszystkie Transfer z tym adresem jako `from`" wciąż działa (frontend hashuje swój adres i porównuje).
- Ale **nie da się odzyskać pełnego adresu z topika** — to strata informacji.

### Opcje

#### Opcja A — topic zostaje 32 B (adres hashowany)

- qrvmone już tak robi
- Zgodność z Ethereum
- Filtrowanie po adresach działa (z hashowaniem)
- Brak odzyskania adresu z topika

#### Opcja B — topic poszerzony do 48 B (pełny adres)

- `types.Log.Topics` → `[]Address48` lub podobnie
- `opLog0..4` zapisuje 48 bajtów per topic ze stosu
- Wymaga zmian w:
  - RLP encoding log
  - Bloom filter (obecnie liczony z 32B topics)
  - Indexing w DB
  - JSON-RPC (wynik filtrów)
  - qrvmc (nowy typ `qrvmc_bytes48`)

#### Opcja C — topic poszerzony do 64 B (pełne słowo ze stosu)

- Symetryczne z Opcja B dla ADR-001.
- Wszystkie skutki jak B, ale 64 zamiast 48.

### Analiza trade-offów

| Aspekt | 32 B (A) | 48 B (B) | 64 B (C) |
|---|---|---|---|
| Rozmiar receipta | min. | +50% per topic | +100% per topic |
| Bloom filter | bez zmian | wymaga redesignu | wymaga redesignu |
| Odzyskanie adresu | nie | tak | tak |
| Zgodność z Ethereum tooling | wysoka | niska | niska |
| qrvmone zmiany | brak | tak | tak |
| Storage chain | min | znaczny wzrost | dramatyczny wzrost |

### Praktyczny problem: bloom filter

Bloom filter to probabilistyczna struktura "czy ten blok zawiera log z tym topicem". Ethereum używa 2048-bitowego bloom filtera per blok, gdzie każdy topic aktywuje 3 bity. Poszerzenie topiku wymaga rewizji funkcji hash-to-bloom — przynajmniej update numerów bitów.

### Dodatkowy problem: `FunctionTy`

W EVM ABI typ `function` = adres (20B) + selector (4B) = 24B. Mieści się w topiku 32B.

W naszym VM: adres 48B + selector 4B = 52B. **Nie mieści się nawet w 48-bajtowym topiku**. W topiku 64B — mieści się.

Pytanie: czy chcemy umożliwić indexed function-type w eventach? Jeśli tak — 64B topic jest jedyną opcją. Jeśli nie (i hashujemy function do 32B) — 32B OK.

### Rekomendacja: **Opcja A (32 B) — spójnie z qrvmone**

**Argumenty:**

1. Topicy są narzędziem **filtrowania**, nie odzyskiwania wartości. Jeśli chcesz pełny adres — masz go w `log.data` (nie-indexed fields) albo w `log.address`.
2. Tool ecosystem (Alchemy, The Graph, Ethers.js) zakłada 32-bajtowe topicy. Zmiana to kaskada.
3. qrvmone już wybrał 32B — retro-migracja to dodatkowa praca.
4. Hashowanie adresu do topika to znana technika (dla string/bytes już jest standardem).

**Wymagana dokumentacja:**
- Jak frontend powinien filtrować po adresie: `topic = keccak256(address[:])`, nie `topic = leftPad(address, 32)`.
- Jakie typy są **hashowane do topika** (dynamic: string, bytes; struktury; tablice; adresy 48B; function types) vs **kopiowane bezpośrednio** (int/uint<=256, bool, bytes1..32).

---

## ADR-003: CREATE2 salt width

### Tło

Są dwa opcody tworzenia kontraktu:

- **CREATE** (0xF0): adres nowego kontraktu = `keccak256(rlp(sender, nonce))[ostatnie_bajty]`. Zależy od nonce, więc adres nieprzewidywalny bez znajomości stanu konta.

- **CREATE2** (0xF5): adres = `keccak256(0xff || sender || salt || keccak256(initcode))[ostatnie_bajty]`. **Nie zależy od nonce.** Dzięki temu można policzyć adres kontraktu przed jego deployem — przydatne dla counterfactual contracts (wallets), channels, upgradeable contracts.

Cechą CREATE2 jest **salt** — dowolna wartość wybrana przez dewelopera, 32 B w EVM. Różne salty → różne adresy dla tego samego initcode.

### Kanoniczny wzór EVM

```
address = keccak512("QRL-CREATE2" || 0xff || sender_address_48B || salt_32B || keccak256(initcode))[ostatnie 48 B]
                                  └┬─┘  └────────┬────────┘  └────┬────┘  └───────────┬─────────┘
                                   1B           48B              32B                  32B
```

Razem wejście do keccak256 = 1 + 20 + 32 + 32 = 85 bajtów. Wyjście = 32 B, bierzemy ostatnie 20 = adres.

### Po co salt?

Salt to "dodatkowy identyfikator" pozwalający deterministycznie różnicować instancje kontraktu. Przykłady:

- **Wallet factory:** user podpisuje message → contract liczy salt = hash(user_pubkey) → deployuje wallet z przewidywalnym adresem. User może wysyłać fundusze na ten adres jeszcze przed deployem.

- **State channels:** dwie strony uzgadniają parametry i liczą salt z hash parametrów → wszyscy znają adres channela.

- **Upgradeable contracts (EIP-1014):** nowa wersja kontraktu z innym salt, ale tym samym initcode → inny adres, łatwa migracja.

### Problem w naszej migracji

Stos jest 512-bit. Salt zdejmowany ze stosu jest 64-bajtowy. Ale:

1. `crypto.CreateAddress2` przyjmuje `salt [32]byte`:
   ```go
   func CreateAddress2(b common.Address, salt [32]byte, inithash []byte) common.Address {
       return keccakToAddress48([]byte{0xff}, b.Bytes(), salt[:], inithash)
   }
   ```

2. `opCreate2` obecnie konwertuje:
   ```go
   salt := scope.Stack.pop()              // uint512 (64 B)
   interpreter.qrvm.Create2(..., bigEndowment, &salt)
   ```
   gdzie `Create2` wewnętrznie wołą `salt.Bytes32()` — obcięcie do 32B.

3. qrvmone robi to samo: `msg.create2_salt = store_uint(salt)` → `qrvmc_bytes32` (32 B).

### Opcje

#### Opcja A — salt 32 B (obecnie, = qrvmone)

Derivacja adresu kontraktu:
```
address_48B = keccak_to_48B(0xff || sender_48B || salt_32B || keccak256(initcode))
```

#### Opcja B — salt 64 B (pełny ze stosu)

```
address_48B = keccak_to_48B(0xff || sender_48B || salt_64B || keccak256(initcode))
```

### Analiza

**Kryptograficznie** — 32B salt daje 256 bitów wyboru, czyli 2^128 ataków urodzinowych na kolizję adresów między różnymi saltami (dla tego samego sender + initcode). To wystarczająco dużo dla praktycznych zastosowań.

**Użyteczność** — salt jest sztuczny; jego rozmiar nie koreluje z naturalnymi danymi. Deweloperzy hashują swoje parametry do 32B obecnie; do 64B również mogliby, ale zysk jest symboliczny.

**Kompatybilność z qrvmone** — Opcja A zero pracy, Opcja B wymaga update.

**Counterfactual address prediction** — frontend liczący adres kontraktu przed deployem wymaga identycznej formuły. Jeśli go-qrl/qrvmone rozjadą się z ecosystem tools, rachowanie adresów zawodzi. Stabilność tu jest kluczowa.

### Rekomendacja: **Opcja A (32 B, = qrvmone)**

Argumenty:
- Brak zysku użyteczności z 64B salt.
- Zgodność z qrvmone.
- Salt jest sztuczny, programista może skompresować dowolne dane do 32B (hashowanie).
- Stabilna, przewidywalna formuła derivacji adresu dla zewnętrznych narzędzi.

### Uwaga — formuła derivacji się i tak zmieniła

Niezależnie od wyboru 32B vs 64B salt, **formuła jako całość** się zmieniła (bo sender jest teraz 48B a wyjście 48B). Więc adres policzony przez CREATE2 w QRL nie jest równy adresowi policzonemu tą samą formułą w Ethereum. **Tool ecosystem i tak wymaga aktualizacji** — to nieuniknione.

---

## ADR-004: Schemat derivacji adresu z klucza publicznego

### Tło

**Adres** — identyfikator konta lub kontraktu. W EVM generowany z klucza publicznego (dla kont zewn. — EOA) albo z hasha CREATE/CREATE2 (dla kontraktów).

Dla EOA klasyczny Ethereum:
```
address = keccak256(secp256k1_pubkey_bytes)[ostatnie 20 B]
```

Wymagania od funkcji derivacji:
1. **Deterministyczna** — ten sam klucz zawsze daje ten sam adres.
2. **Odporna na kolizje** — trudno znaleźć dwa różne klucze dające ten sam adres.
3. **Jednokierunkowa** — z adresu nie da się odzyskać klucza (prywatność, odporność na ataki).
4. **Szybka** — liczona przy każdej weryfikacji podpisu.

Dla adresu 20 B (160 bitów):
- Kolizje: urodzinowo 2^80 → **marginalne bezpieczeństwo** (dziś akademicki próg to 2^128).
- Pre-image: 2^160 → OK.

Dla adresu 48 B (384 bitów) — mamy znacznie więcej miejsca; byłoby nierozsądne marnować je słabą derivacją.

### Bezpieczeństwo hash-to-N

Ogólna reguła — jeśli hash H ma wyjście M bitów, a my bierzemy z niego N bitów (N < M), to:
- Kolizja przez paradoks urodzinowy: **2^(N/2)**
- Pre-image: **2^N**

Więc dla adresu 48 B = 384 bitów: maksymalne możliwe bezpieczeństwo = 2^192 kolizje, 2^384 pre-image. **Do tego powinniśmy dążyć.**

### QRL kontekst: ML-DSA-87

QRL używa podpisów **post-kwantowych** — algorytm **ML-DSA-87** (jeden z NIST finalists, rozmiar pubkey ~2.5 KB, podpis ~4.5 KB).

Generalna wskazówka dla derivacji adresu z ML-DSA-87 pubkey: traktujemy pubkey jako surowy bajtowy materiał i hashujemy. ML-DSA ma pseudolosowe pubkey, więc dowolny dobry hash daje jednorodne wyjście.

W go-qrl są dwie ścieżki derivacji adresu:

1. **ECDSA secp256k1 (legacy, testowe)** — `crypto.PubkeyToAddress(ecdsa.PublicKey)`. Klucz SEC1 uncompressed (65B z prefiksem 0x04), hashujemy 64 bajty (X, Y).
2. **ML-DSA-87 (produkcyjne)** — `walletmldsa87.Wallet.GetAddress()` w go-qrllib. Szczegóły derivacji są w qrllib, nie w go-qrl.

**ADR-004 dotyczy głównie ścieżki (2)**, ale spójność wymaga jednakowego schematu dla obu.

### Opcje

#### Opcja A — kompozycja dwóch Keccak256 (obecnie)

```go
func keccakToAddress48(data ...[]byte) common.Address {
    h1 := Keccak256(data...)         // 32 B
    h2 := Keccak256(h1)              // 32 B
    return common.BytesToAddress(append(h1, h2[:16]...))  // 32 + 16 = 48 B
}
```

**Analiza bezpieczeństwa:**

Adres = `H(pk) || H(H(pk))[:16]`, gdzie H = Keccak256.

Kolizja adresu wymaga:
- `H(pk1) == H(pk2)` (pierwsze 32 B zgadza się)
- **I** `H(H(pk1))[:16] == H(H(pk2))[:16]`

Ale jeśli `H(pk1) == H(pk2)` to automatycznie `H(H(pk1)) == H(H(pk2))` (determinizm H). Więc drugi warunek jest darmowy gdy pierwszy się spełnia.

→ **Odporność na kolizje = 2^128** (to co daje `H(pk)`).

Dodatkowe 16 bajtów nie przynosi bezpieczeństwa. **Wąskie gardło to Keccak256.**

#### Opcja B — Keccak512 z truncation

```go
func keccakToAddress48(data ...[]byte) common.Address {
    h := Keccak512(data...)        // 64 B
    return common.BytesToAddress(h[16:])  // ostatnie 48 B
}
```

**Analiza:** wyjście Keccak512 (64 B) jest jednorodnie losowe. Bierzemy ostatnie 48 B. Kolizja adresów wymaga kolizji 384 bitów wyjścia.

→ **Odporność na kolizje = 2^192.**

**2^64 razy lepsze niż Opcja A.**

#### Opcja C — dedykowana derivacja z ML-DSA

Na przykład:
```
address = DomainHash("QRL-ADDR", T_0 || T_1 || ...)
```
gdzie T_i to konkretne komponenty struktury pubkey ML-DSA.

**Zysk:** zero kryptograficzny — ML-DSA pubkey już jest pseudolosowy. **Koszt:** dużo pracy projektowej + ryzyko że przy zmianie ML-DSA parameters adresy się zmienią.

#### Opcja B+ — Keccak512 + separacja domen

```go
func keccakToAddress48(data ...[]byte) common.Address {
    prefixed := append([][]byte{[]byte("QRL-ADDR-v1")}, data...)
    h := Keccak512(prefixed...)
    return common.BytesToAddress(h[16:])
}
```

**Po co domain separation:** zabezpiecza przed cross-protocol collisions. Jeśli gdzieś indziej hashujemy podobne dane Keccak512 (np. block hash), ktoś mógłby próbować znaleźć input który jednocześnie generuje valid adres i valid block-hash-like wartość. Prefix "QRL-ADDR-v1" blokuje tę klasę ataków.

Koszt: 11 dodatkowych bajtów na wejściu hasha (pomijalne). Zysk: pewność obronna.

### Porównanie

| Schemat | Wyjście | Kolizje | Pre-image | Domain sep. | qrvmone/qrvmc zmiany |
|---|---|---|---|---|---|
| Stary EVM (20 B) | `K256[12:]` | 2^80 | 2^160 | brak | — |
| A (obecne) | `K256 ∥ K256²[:16]` | **2^128** | 2^160 | brak | brak |
| B | `K512[16:]` | 2^192 | 2^384 | brak | brak |
| B+ (rekomendacja) | `K512("QRL-ADDR-v1" ∥ pk)[16:]` | 2^192 | 2^384 | **tak** | brak |
| C | dedykowane | zależy | zależy | tak (gratis) | brak |

### Rekomendacja: **Opcja B+**

**Argumenty:**
1. **Kryptograficznie mocniejsza** — pełne 192 bity kolizji vs 128 w obecnej implementacji.
2. **Jedna wywołanie hasha** — prostsze, łatwiejsze do formalnej analizy.
3. **Keccak512 już dostępny w kodzie** i pasuje do 48-bajtowego wyniku adresu.
4. **Separacja domen** zabezpiecza przed klasą ataków cross-protocol.
5. **Bez zmian w qrvmone/qrvmc** — derivacja adresu dzieje się poza VM.

**Implementacja zmiany:**
- `crypto/crypto.go`: `keccakToAddress48` — zmiana ~10 linii.
- `crypto/crypto_test.go`: regeneracja 5 fixtur (oczekiwane adresy).
- **Zero zmian w qrvmone** — sender adres przychodzi z hosta, VM go nie oblicza.

**W go-qrllib (jeśli koordynujemy):** identyczny schemat dla ML-DSA-87. 

### Warianty po decyzji

Gdy ADR-004 jest DONE:
- Update `crypto/crypto.go`.
- Update `go-qrllib/wallet/ml_dsa_87/wallet.go` (zewnętrznie).
- Wywalić TODO komentarz w `crypto/pqcrypto/wallet/wallet.go`.
- Regenerować fixtury w `crypto_test.go`, `core/state/state_test.go`, genesis hashes.

---

## ADR-005: Rekalibracja modelu gazu

### Tło

**Gas** to jednostka kosztu wykonania w VM. Zabezpiecza przed:
- Nieskończonymi pętlami (transakcja ma limit gazu).
- DoS (każdy opcod kosztuje, spam wymaga pieniędzy).
- Nieproporcjonalnym zużyciem zasobów (CPU, RAM, disk IO).

Każdy opcod ma zdefiniowany koszt. Przykłady (EVM 256-bit):
- `ADD`, `SUB`, `AND`, `OR`, `XOR`: 3 gazu (proste ALU)
- `MUL`, `DIV`: 5 gazu (trochę droższe)
- `EXP`: 10 + 50 per byte exponent (dynamiczny)
- `SLOAD`: 2100 gazu cold / 100 gazu warm (I/O disk)
- `SSTORE`: ~20000 gazu na zapis niezerowej wartości (ciężki I/O + state tree update)
- `KECCAK256`: 30 + 6 per 32-byte word (obliczenia + kopiowanie)
- **Memory expansion**: kwadratowy koszt, `3 * N + N²/512` dla N słów (gdzie 1 słowo = 32 B)

### Problem

Model gazu w EVM został skalibrowany dla:
- 32-bajtowego słowa (stack, memory)
- 32-bajtowego slotu storage
- 20-bajtowego adresu
- określonych kosztów sprzętu ~2015/2020

W naszym VM:
- Słowo stack/memory: **64 B** (2× większe)
- Adres: **48 B** (2.4× większy)
- Słowo storage: 32 B (bez zmian — ADR-001)

### Co się zmienia w kosztach naturalnie?

Kilka opcodów **naturalnie** ma inne koszty w 64-bit VM:
- **MLOAD/MSTORE**: pracują na 64B zamiast 32B. Kopiowanie 2× większe, ale operacja dalej O(1). Bazowy koszt 3 gazu powinien zostać, ale memory expansion model wymaga update.
- **Memory expansion**: obecnie `(size+31)/32 * 32` → `(size+63)/64 * 64`. Słowa liczymy w 64B jednostkach. Koszt kwadratowy w słowach — jeśli chcemy zachować "mniej-więcej takie same koszty w bajtach" to wzór trzeba przeliczyć.
- **KECCAK256**: koszt per-word. Input dalej bajtowy, więc logicznie powinno zostać "per 32 bytes of input". Nie zmienia się z rozmiarem słowa VM.
- **CALLDATALOAD**: pobiera 32 B (w EVM) / 64 B (u nas). 2× więcej danych per opcod. Bazowy koszt powinien wzrosnąć.
- **CALLDATACOPY, CODECOPY, RETURNDATACOPY, EXTCODECOPY**: koszt per-word. Liczenie słów w 64B jednostkach.
- **SHL/SHR/SAR**: działają na 512-bit wartości zamiast 256-bit. Marginalnie droższe, ale dalej O(1). Prawdopodobnie zostawiamy bazowy koszt.
- **MUL/DIV/EXP**: uint512 arytmetyka ~2-4× wolniejsza niż uint256. Koszty powinny wzrosnąć — ale o ile?

### Memory cost formula

Klasyczny EVM:
```
memory_cost(N_words) = 3 * N + N² / 512
```
gdzie N = liczba 32-bajtowych słów.

W naszym VM:
- Opcja A: zachowujemy `N = liczba 64-bajtowych słów`. Wzór ten sam, ale N jest 2× mniejsze dla tej samej ilości bajtów. Koszt w bajtach = połowa oryginalnego EVM. **Pamięć jest relatywnie tańsza.**
- Opcja B: `N = liczba 32-bajtowych słów` (jak EVM), ale pamięć jest alokowana w 64-bajtowych porcjach. Koszt identyczny jak EVM.
- Opcja C: rekalibracja współczynników tak, żeby koszt per bajt był stały.

### Obliczenia (przykład)

64 KB pamięci:
- EVM: N = 2048 słów 32B. Cost = 3×2048 + 2048²/512 = 6144 + 8192 = **14336 gazu**.
- Nasz VM (Opcja A): N = 1024 słów 64B. Cost = 3×1024 + 1024²/512 = 3072 + 2048 = **5120 gazu**. 2.8× tańsze.
- Nasz VM (Opcja B): N = 2048 słów 32B (liczymy tak jak EVM). Cost = 14336.
- Nasz VM (Opcja C, rekalibracja do "per-bajt"): trzeba dobrać nowy współczynnik kwadratowy. Np. `N² / 128` gdzie N = 1024 → `1024²/128 = 8192`. Razem 3072 + 8192 = 11264. Bliższe EVM.

### Opcody uint512 — ile wolniej?

Benchmark hipotetyczny (uint256 holiman vs uint512 math/big):
- Add: 20ns vs 80ns (~4×)
- Mul: 50ns vs 400ns (~8×)
- Exp: 1µs vs 10µs (~10×)

Po migracji na natywny uint512 (Faza 1.2): ~2× wolniej niż uint256.

**Rekomendacja:** stałe koszty typu `GasFastestStep=3` zostawić; dynamiczne koszty (per-word dla KECCAK, MUL-mod-exp, precompile) rekalibrować wg real-world benchmarków.

### Opcje ADR-005

#### Opcja A — minimalne skalowanie (obecne)

Tylko to co konieczne dla kompilacji:
- `toWordSize * 32 → * 64` (zrobione)
- `newMemSizeWords * 32 → * 64` w `memoryGasCost` (zrobione)

Reszta: bez zmian. Pamięć efektywnie tańsza, niektóre opcody też.

#### Opcja B — pełna rekalibracja ekonomiczna

- Benchmark realistyczny każdego opcodu.
- Nowa tabela kosztów zgodnie z rzeczywistym CPU/IO.
- Aktualizacja memory cost formula.
- Możliwe nowe stałe: `GasSlowerStep512` itp.

#### Opcja C — utrzymanie EVM-compatible kosztów per-bajt

- Koszt memory w bajtach = taki sam jak EVM (więc wzór liczony w 32B jednostkach).
- Koszt computational ×1 (stałe jak EVM).
- Cel: tani migration path dla istniejących kontraktów Solidity.

### Rekomendacja: **Opcja B** (rekalibracja), ale z fazami

1. **Faza 1 (teraz)**: zatrzymać się na A — minimalne skalowanie dla poprawności. To aktualny stan.
2. **Faza 2 (przed testnet v3)**: dodać benchmarki opcodów, audytowalną tabelę "stary koszt → nowy koszt", decyzję per-opcod.
3. **Faza 3 (przed mainnet)**: ekonomista + auditor kryptoekonomiczny ocenia czy koszty są DoS-resistant.

### Konsekwencje dla testów

Większość testów z `gasUsed` jako expected wartość wymaga regeneracji. To **największe ryzyko fixtur** — gas jest wszędzie.

---

## ADR-006: ChainID dla 48B sieci

### Tło

**ChainID** to unikalna liczba identyfikująca blockchain. Używana w:
- **EIP-155 replay protection** — podpis transakcji zawiera ChainID, więc transakcja z jednej sieci nie może być odtworzona na innej.
- **Wallet UX** — MetaMask pokazuje "connected to chain 137" (Polygon) vs "chain 1" (Ethereum mainnet).

QRL ma obecnie:
- Testnet v2: ChainID = jakiś konkretny numer (do sprawdzenia w `params/config.go`)
- Mainnet: jeszcze nie uruchomiony, ale planowany ChainID

### Problem

Jeśli wypuścimy 48B-adresową sieć pod **tym samym** ChainID co 20B-adresowa:

1. Transakcje skonstruowane dla 20B sieci mogą być odtworzone w 48B sieci (jeśli signer podpisuje ten sam materiał). → **Replay attack.**
2. Użytkownicy portfela myślą że są na "tej samej sieci" — podpisują transakcje które nie znaczą tego co myślą.
3. Eksploratory, bootnode'y dostają niespójne dane pod tym samym identyfikatorem.

### Opcje

#### Opcja A — Nowy ChainID dla każdej wersji

- Testnet v3 (48B): nowy ChainID (np. poprzedni + 1).
- Mainnet (48B, nowy): nowy ChainID.
- Explicit break z poprzednimi sieciami.

#### Opcja B — Reuse ChainID

- Testnet v3 = testnet v2 ChainID.
- Argument: "testnet i tak nie ma wartości, replay nie jest groźny".

### Analiza

**Argumenty za A:**
- Replay protection (standardowa praktyka EIP-155).
- Jasna separacja w wallet UX.
- Brak miksowania danych w eksploratorach.

**Argumenty za B:**
- Mniej aktualizacji w konfigach bootnode'ów itp.
- Testnet i tak jest volatile.

### Rekomendacja: **Opcja A — nowy ChainID dla sieci 48B**

**Zdecydowanie A.** Każda zmiana mogąca wygenerować niekompatybilne transakcje = nowy ChainID. To punkt higieny kryptograficznej i UX, nie technikalia.

---

## ADR-007: MaxInitCodeSize (EIP-3860)

### Tło

**EIP-3860** (Shanghai hardfork) wprowadza limit rozmiaru inicjalizacyjnego kodu kontraktu:
```
max_initcode_size = 2 * max_contract_code_size = 2 * 24576 = 49152 bajty
```

Po co limit?
- Init code jest analizowany przy każdym deployu (JUMPDEST analysis = O(n)).
- Bez limitu ataker może wkleić ogromny init code i zmusić walidatorów do dużego CPU.
- Gas samego deployu jest mały (w porównaniu do analizy), więc limit dopełnia ekonomię ataku.

### Problem

Limit 49152 = 2 * 24576 był skalibrowany dla:
- 32-bajtowe słowo VM
- Specyficzny koszt analizy JUMPDEST

W 64-bajtowym VM:
- JUMPDEST analysis skanuje kod opcod-po-opcodzie, bitmap 1-bitowy per bajt kodu. **Koszt analizy = O(n) gdzie n = bajty kodu.** Nie zmienia się z word size.
- `toWordSize` zmieniło się z `/32` na `/64`. Liczba "słów kodu" dla tego samego kodu = połowa.

### Opcje

#### Opcja A — limit pozostaje 49152 bajtów

Argumenty:
- Rozmiar kodu w bajtach się nie zmienił (nie wydłużamy kontraktów dlatego że słowo VM jest szersze).
- Koszt analizy JUMPDEST per-bajt taki sam.
- Konsystencja z EVM.

#### Opcja B — limit 2× = 98304 bajty

Argumenty:
- Symetria z innymi podwojeniami.
- Kontrakty z wieloma 64-bajtowymi stałymi (PUSH64) mogą naturalnie być większe.

#### Opcja C — limit dobrany empirycznie

- Benchmark: ile kod mogą walidatorzy przeanalizować w rozsądnym czasie?
- Na podstawie benchmarków ustalić nowy limit.

### Analiza

**PUSH64 vs PUSH32** — pojawiła się nowa instrukcja PUSH64, która zajmuje **65 bajtów** w kodzie (1 opcode + 64 immediate). Kontrakt który "kiedyś" używał PUSH32 (33 bajty) teraz używa PUSH64 (65 bajtów) — to **2× więcej kodu** dla tej samej semantyki.

Jeśli zostawimy limit 49152, kontrakty używające intensywnie PUSH64 szybciej uderzają w limit.

**Ale:** limit EIP-3860 był dobrany konserwatywnie. 2× zapas to ~98KB, które dalej jest analyzowalne w sensownym czasie.

### Rekomendacja: **Opcja B — 98304 bajty (2×)**

**Argumenty:**
- Kontrakty z PUSH64 naturalnie są większe bez wzrostu złożoności obliczeniowej.
- Analiza JUMPDEST jest O(n) w bajtach — 2× większy limit to 2× więcej CPU przy analizie, wciąż rozsądne.
- Użytkownik który kompiluje Solidity-with-48B-addresses dostaje podobną zdolność ekspresji co EVM programmer.

**Walidacja:** benchmark pokazujący że analiza 98KB kodu < 10ms na typowej maszynie walidatora.

---

## Podsumowanie rekomendacji i zależności

| ADR | Tytuł | Rekomendacja | Zmienia C++? |
|---|---|---|---|
| 001 | Storage slot width | 32 B | Nie (qrvmone już tak) |
| 002 | Log topic width | 32 B | Nie (qrvmone już tak) |
| 003 | CREATE2 salt width | 32 B | Nie (qrvmone już tak) |
| 004 | Address derivation | Keccak512 + domain sep | Nie (derivacja poza VM) |
| 005 | Gas recalibration | Faza B (pełna rekalibracja) | Tak, synchronicznie |
| 006 | ChainID | Nowy dla 48B | Nie (runtime value) |
| 007 | MaxInitCodeSize | 2× = 98304 B | Zależy od hosta |

### Uwagi końcowe

- **ADR-001, 002, 003** jeśli rekomendowane (32B) → **zero zmian w qrvmone/qrvmc**. Faza 2 sprowadza się do udokumentowania wyboru.
- **ADR-004** → tylko zmiana w Go (i zewn. qrllib).
- **ADR-005** → iteracyjny proces, wymaga benchmarków.
- **ADR-006, 007** → proste decyzje konfiguracyjne.

Jeśli wszystkie rekomendacje przyjęte, to największy blok pracy z ADR-ów to:
1. Rekalibracja gazu (ADR-005) — kilka tygodni.
2. Zmiana derivacji adresu (ADR-004) — dzień + regeneracja fixtur.
