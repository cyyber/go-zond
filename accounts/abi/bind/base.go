// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package bind

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"sync"

	qrl "github.com/theQRL/go-qrl"
	"github.com/theQRL/go-qrl/accounts/abi"
	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/go-qrl/core/types"
	"github.com/theQRL/go-qrl/crypto"
	"github.com/theQRL/go-qrl/event"
)

const basefeeWiggleMultiplier = 2

var (
	errNoEventSignature       = errors.New("no event signature")
	errEventSignatureMismatch = errors.New("event signature mismatch")
)

// SignerFn is a signer function callback when a contract requires a method to
// sign the transaction before submission.
type SignerFn func(common.Address, *types.Transaction) (*types.Transaction, error)

// CallOpts is the collection of options to fine tune a contract call request.
type CallOpts struct {
	Pending     bool            // Whether to operate on the pending state or the last known one
	From        common.Address  // Optional the sender address, otherwise the first account is used
	BlockNumber *big.Int        // Optional the block number on which the call should be performed
	Context     context.Context // Network context to support cancellation and timeouts (nil = no timeout)
}

// TransactOpts is the collection of authorization data required to create a
// valid QRL transaction.
type TransactOpts struct {
	From   common.Address // QRL account to send the transaction from
	Nonce  *big.Int       // Nonce to use for the transaction execution (nil = use pending state)
	Signer SignerFn       // Method to use for signing the transaction (mandatory)

	Value     *big.Int // Funds to transfer along the transaction (nil = 0 = no funds)
	GasFeeCap *big.Int // Gas fee cap to use for the 1559 transaction execution (nil = gas price oracle)
	GasTipCap *big.Int // Gas priority fee cap to use for the 1559 transaction execution (nil = gas price oracle)
	GasLimit  uint64   // Gas limit to set for the transaction execution (0 = estimate)

	Context context.Context // Network context to support cancellation and timeouts (nil = no timeout)

	NoSend bool // Do all transact steps but do not send the transaction
}

// FilterOpts is the collection of options to fine tune filtering for events
// within a bound contract.
type FilterOpts struct {
	Start uint64  // Start of the queried range
	End   *uint64 // End of the range (nil = latest)

	Context context.Context // Network context to support cancellation and timeouts (nil = no timeout)
}

// WatchOpts is the collection of options to fine tune subscribing for events
// within a bound contract.
type WatchOpts struct {
	Start   *uint64         // Start of the queried range (nil = latest)
	Context context.Context // Network context to support cancellation and timeouts (nil = no timeout)
}

// MetaData collects all metadata for a bound contract.
type MetaData struct {
	mu   sync.Mutex
	Sigs map[string]string
	Bin  string
	ABI  string
	ab   *abi.ABI
}

func (m *MetaData) GetAbi() (*abi.ABI, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ab != nil {
		return m.ab, nil
	}
	if parsed, err := abi.JSON(strings.NewReader(m.ABI)); err != nil {
		return nil, err
	} else {
		m.ab = &parsed
	}
	return m.ab, nil
}

// BoundContract is the base wrapper object that reflects a contract on the
// QRL network. It contains a collection of methods that are used by the
// higher level contract bindings to operate.
type BoundContract struct {
	address    common.Address     // Deployment address of the contract on the QRL blockchain
	abi        abi.ABI            // Reflect based ABI to access the correct QRL methods
	caller     ContractCaller     // Read interface to interact with the blockchain
	transactor ContractTransactor // Write interface to interact with the blockchain
	filterer   ContractFilterer   // Event filtering to interact with the blockchain
}

// NewBoundContract creates a low level contract interface through which calls
// and transactions may be made through.
func NewBoundContract(address common.Address, abi abi.ABI, caller ContractCaller, transactor ContractTransactor, filterer ContractFilterer) *BoundContract {
	return &BoundContract{
		address:    address,
		abi:        abi,
		caller:     caller,
		transactor: transactor,
		filterer:   filterer,
	}
}

// DeployContract deploys a contract onto the QRL blockchain and binds the
// deployment address with a Go wrapper.
func DeployContract(opts *TransactOpts, abi abi.ABI, bytecode []byte, backend ContractBackend, params ...any) (common.Address, *types.Transaction, *BoundContract, error) {
	// Otherwise try to deploy the contract
	c := NewBoundContract(common.Address{}, abi, backend, backend, backend)

	input, err := c.abi.Pack("", params...)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	tx, err := c.transact(opts, nil, append(bytecode, input...))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	c.address = crypto.CreateAddress(opts.From, tx.Nonce())
	return c.address, tx, c, nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (c *BoundContract) Call(opts *CallOpts, results *[]any, method string, params ...any) error {
	// Don't crash on a lazy user
	if opts == nil {
		opts = new(CallOpts)
	}
	if results == nil {
		results = new([]any)
	}
	// Pack the input, call and unpack the results
	input, err := c.abi.Pack(method, params...)
	if err != nil {
		return err
	}
	var (
		msg    = qrl.CallMsg{From: opts.From, To: &c.address, Data: input}
		ctx    = ensureContext(opts.Context)
		code   []byte
		output []byte
	)
	if opts.Pending {
		pb, ok := c.caller.(PendingContractCaller)
		if !ok {
			return ErrNoPendingState
		}
		output, err = pb.PendingCallContract(ctx, msg)
		if err != nil {
			return err
		}
		if len(output) == 0 {
			// Make sure we have a contract to operate on, and bail out otherwise.
			if code, err = pb.PendingCodeAt(ctx, c.address); err != nil {
				return err
			} else if len(code) == 0 {
				return ErrNoCode
			}
		}
	} else {
		output, err = c.caller.CallContract(ctx, msg, opts.BlockNumber)
		if err != nil {
			return err
		}
		if len(output) == 0 {
			// Make sure we have a contract to operate on, and bail out otherwise.
			if code, err = c.caller.CodeAt(ctx, c.address, opts.BlockNumber); err != nil {
				return err
			} else if len(code) == 0 {
				return ErrNoCode
			}
		}
	}

	if len(*results) == 0 {
		res, err := c.abi.Unpack(method, output)
		*results = res
		return err
	}
	res := *results
	return c.abi.UnpackIntoInterface(res[0], method, output)
}

// Transact invokes the (paid) contract method with params as input values.
func (c *BoundContract) Transact(opts *TransactOpts, method string, params ...any) (*types.Transaction, error) {
	// Otherwise pack up the parameters and invoke the contract
	input, err := c.abi.Pack(method, params...)
	if err != nil {
		return nil, err
	}
	// todo(rjl493456442) check the method is payable or not,
	// reject invalid transaction at the first place
	return c.transact(opts, &c.address, input)
}

// RawTransact initiates a transaction with the given raw calldata as the input.
// It's usually used to initiate transactions for invoking **Fallback** function.
func (c *BoundContract) RawTransact(opts *TransactOpts, calldata []byte) (*types.Transaction, error) {
	// todo(rjl493456442) check the method is payable or not,
	// reject invalid transaction at the first place
	return c.transact(opts, &c.address, calldata)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (c *BoundContract) Transfer(opts *TransactOpts) (*types.Transaction, error) {
	// todo(rjl493456442) check the payable fallback or receive is defined
	// or not, reject invalid transaction at the first place
	return c.transact(opts, &c.address, nil)
}

func (c *BoundContract) createDynamicTx(opts *TransactOpts, contract *common.Address, input []byte, head *types.Header) (*types.Transaction, error) {
	// Normalize value
	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	// Estimate TipCap
	gasTipCap := opts.GasTipCap
	if gasTipCap == nil {
		tip, err := c.transactor.SuggestGasTipCap(ensureContext(opts.Context))
		if err != nil {
			return nil, err
		}
		gasTipCap = tip
	}
	// Estimate FeeCap
	gasFeeCap := opts.GasFeeCap
	if gasFeeCap == nil {
		gasFeeCap = new(big.Int).Add(
			gasTipCap,
			new(big.Int).Mul(head.BaseFee, big.NewInt(basefeeWiggleMultiplier)),
		)
	}
	if gasFeeCap.Cmp(gasTipCap) < 0 {
		return nil, fmt.Errorf("maxFeePerGas (%v) < maxPriorityFeePerGas (%v)", gasFeeCap, gasTipCap)
	}
	// Estimate GasLimit
	gasLimit := opts.GasLimit
	if opts.GasLimit == 0 {
		var err error
		gasLimit, err = c.estimateGasLimit(opts, contract, input, gasTipCap, gasFeeCap, value)
		if err != nil {
			return nil, err
		}
	}
	// create the transaction
	nonce, err := c.getNonce(opts)
	if err != nil {
		return nil, err
	}
	baseTx := &types.DynamicFeeTx{
		To:        contract,
		Nonce:     nonce,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Gas:       gasLimit,
		Value:     value,
		Data:      input,
	}
	return types.NewTx(baseTx), nil
}

func (c *BoundContract) estimateGasLimit(opts *TransactOpts, contract *common.Address, input []byte, gasTipCap, gasFeeCap, value *big.Int) (uint64, error) {
	if contract != nil {
		// Gas estimation cannot succeed without code for method invocations.
		if code, err := c.transactor.PendingCodeAt(ensureContext(opts.Context), c.address); err != nil {
			return 0, err
		} else if len(code) == 0 {
			return 0, ErrNoCode
		}
	}
	msg := qrl.CallMsg{
		From:      opts.From,
		To:        contract,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Value:     value,
		Data:      input,
	}
	return c.transactor.EstimateGas(ensureContext(opts.Context), msg)
}

func (c *BoundContract) getNonce(opts *TransactOpts) (uint64, error) {
	if opts.Nonce == nil {
		return c.transactor.PendingNonceAt(ensureContext(opts.Context), opts.From)
	} else {
		return opts.Nonce.Uint64(), nil
	}
}

// transact executes an actual transaction invocation, first deriving any missing
// authorization fields, and then scheduling the transaction for execution.
func (c *BoundContract) transact(opts *TransactOpts, contract *common.Address, input []byte) (*types.Transaction, error) {
	// Create the transaction
	var (
		rawTx *types.Transaction
		err   error
	)
	if opts.GasFeeCap != nil && opts.GasTipCap != nil {
		rawTx, err = c.createDynamicTx(opts, contract, input, nil)
	} else {
		if head, errHead := c.transactor.HeaderByNumber(ensureContext(opts.Context), nil); errHead != nil {
			return nil, errHead
		} else {
			rawTx, err = c.createDynamicTx(opts, contract, input, head)
		}
	}
	if err != nil {
		return nil, err
	}
	// Sign the transaction and schedule it for execution
	if opts.Signer == nil {
		return nil, errors.New("no signer to authorize the transaction with")
	}
	signedTx, err := opts.Signer(opts.From, rawTx)
	if err != nil {
		return nil, err
	}
	if opts.NoSend {
		return signedTx, nil
	}
	if err := c.transactor.SendTransaction(ensureContext(opts.Context), signedTx); err != nil {
		return nil, err
	}
	return signedTx, nil
}

// FilterLogs filters contract logs for past blocks, returning the necessary
// channels to construct a strongly typed bound iterator on top of them.
func (c *BoundContract) FilterLogs(opts *FilterOpts, name string, query ...[]any) (chan types.Log, event.Subscription, error) {
	// Don't crash on a lazy user
	if opts == nil {
		opts = new(FilterOpts)
	}
	eventDef, ok := c.abi.Events[name]
	if !ok {
		return nil, nil, fmt.Errorf("event %q not found", name)
	}
	if err := validateEventQuery(eventDef, query); err != nil {
		return nil, nil, err
	}
	// Non-anonymous events start with their signature. Anonymous events omit the
	// selector, so their first query rule applies to their first indexed input.
	if !eventDef.Anonymous {
		query = append([][]any{{eventDef.ID}}, query...)
	}

	topics, err := abi.MakeTopics(query...)
	if err != nil {
		return nil, nil, err
	}
	// Start the background filtering
	logs := make(chan types.Log, 128)

	config := qrl.FilterQuery{
		Addresses: []common.Address{c.address},
		Topics:    topics,
		FromBlock: new(big.Int).SetUint64(opts.Start),
	}
	if opts.End != nil {
		config.ToBlock = new(big.Int).SetUint64(*opts.End)
	}
	/* TODO(karalabe): Replace the rest of the method below with this when supported
	sub, err := c.filterer.SubscribeFilterLogs(ensureContext(opts.Context), config, logs)
	*/
	buff, err := c.filterer.FilterLogs(ensureContext(opts.Context), config)
	if err != nil {
		return nil, nil, err
	}
	sub, err := event.NewSubscription(func(quit <-chan struct{}) error {
		for _, log := range buff {
			select {
			case logs <- log:
			case <-quit:
				return nil
			}
		}
		return nil
	}), nil

	if err != nil {
		return nil, nil, err
	}
	return logs, sub, nil
}

// WatchLogs filters subscribes to contract logs for future blocks, returning a
// subscription object that can be used to tear down the watcher.
func (c *BoundContract) WatchLogs(opts *WatchOpts, name string, query ...[]any) (chan types.Log, event.Subscription, error) {
	// Don't crash on a lazy user
	if opts == nil {
		opts = new(WatchOpts)
	}
	eventDef, ok := c.abi.Events[name]
	if !ok {
		return nil, nil, fmt.Errorf("event %q not found", name)
	}
	if err := validateEventQuery(eventDef, query); err != nil {
		return nil, nil, err
	}
	// Non-anonymous events start with their signature. Anonymous events omit the
	// selector, so their first query rule applies to their first indexed input.
	if !eventDef.Anonymous {
		query = append([][]any{{eventDef.ID}}, query...)
	}

	topics, err := abi.MakeTopics(query...)
	if err != nil {
		return nil, nil, err
	}
	// Start the background filtering
	logs := make(chan types.Log, 128)

	config := qrl.FilterQuery{
		Addresses: []common.Address{c.address},
		Topics:    topics,
	}
	if opts.Start != nil {
		config.FromBlock = new(big.Int).SetUint64(*opts.Start)
	}
	sub, err := c.filterer.SubscribeFilterLogs(ensureContext(opts.Context), config, logs)
	if err != nil {
		return nil, nil, err
	}
	return logs, sub, nil
}

// UnpackLog unpacks a retrieved log into the provided output structure.
func (c *BoundContract) UnpackLog(out any, event string, log types.Log) error {
	eventDef, topics, err := c.eventTopics(event, log.Topics)
	if err != nil {
		return err
	}
	nonIndexed := eventDef.Inputs.NonIndexed()
	if len(nonIndexed) > 0 {
		values, err := nonIndexed.Unpack(log.Data)
		if err != nil {
			return err
		}
		if err := copyEventValues(out, eventDef.Inputs, false, values); err != nil {
			return err
		}
	}
	var indexed abi.Arguments
	for _, arg := range eventDef.Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	return copyEventTopics(out, eventDef.Inputs, indexed, topics)
}

// UnpackLogIntoMap unpacks a retrieved log into the provided map.
func (c *BoundContract) UnpackLogIntoMap(out map[string]any, event string, log types.Log) error {
	eventDef, topics, err := c.eventTopics(event, log.Topics)
	if err != nil {
		return err
	}
	if out == nil {
		return errors.New("abi: cannot unpack event into a nil map")
	}
	if err := validateEventMapKeys(eventDef); err != nil {
		return err
	}
	nonIndexed := eventDef.Inputs.NonIndexed()
	if len(nonIndexed) > 0 {
		if err := nonIndexed.UnpackIntoMap(out, log.Data); err != nil {
			return err
		}
	}
	var indexed abi.Arguments
	for _, arg := range eventDef.Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	return abi.ParseTopicsIntoMap(out, indexed, topics)
}

// validateEventQuery checks typed filter values before MakeTopics erases their
// ABI context. Generated bindings use *big.Int for integer widths without an
// exact Go scalar type, so relying on MakeTopics alone would accept values that
// fit a VM64 topic but not the event's declared width (including negative uints).
// Explicit common.Hash and common.LogTopic values remain supported as raw topic
// rules, which is also how generated bindings query indexed composite values.
func validateEventQuery(eventDef abi.Event, query [][]any) error {
	if err := eventDef.Validate(); err != nil {
		return err
	}
	indexed := make(abi.Arguments, 0, len(eventDef.Inputs))
	for _, argument := range eventDef.Inputs {
		if argument.Indexed {
			indexed = append(indexed, argument)
		}
	}
	if len(query) > len(indexed) {
		return fmt.Errorf("abi: event %q has %d indexed arguments, got %d topic rules", eventDef.RawName, len(indexed), len(query))
	}
	for queryIndex, alternatives := range query {
		argument := indexed[queryIndex]
		for alternativeIndex, rule := range alternatives {
			switch rule.(type) {
			case common.Hash, common.LogTopic:
				continue // Explicit precomputed topic.
			}
			if _, err := (abi.Arguments{{Name: argument.Name, Type: argument.Type}}).Pack(rule); err != nil {
				return fmt.Errorf("abi: invalid topic rule %d for indexed event argument %q (alternative %d): %w", queryIndex, argument.Name, alternativeIndex, err)
			}
		}
	}
	return nil
}

// validateEventMapKeys rejects event schemas which a map cannot represent
// without losing information. Struct decoding remains available for duplicate
// inputs because it can pair fields by occurrence and ABI tag. NewEvent gives
// unnamed inputs stable argN names, so those remain valid map keys.
func validateEventMapKeys(eventDef abi.Event) error {
	seen := make(map[string]struct{}, len(eventDef.Inputs))
	for _, argument := range eventDef.Inputs {
		name := argument.Name
		if _, exists := seen[name]; exists {
			return fmt.Errorf("abi: duplicate event argument name %q cannot be unpacked into a map", name)
		}
		seen[name] = struct{}{}
	}
	return nil
}

// eventTopics resolves an ABI event and removes the signature topic for
// non-anonymous events. Anonymous events use every log topic as indexed data.
func (c *BoundContract) eventTopics(name string, topics []common.LogTopic) (abi.Event, []common.LogTopic, error) {
	eventDef, ok := c.abi.Events[name]
	if !ok {
		return abi.Event{}, nil, fmt.Errorf("event %q not found", name)
	}
	if err := eventDef.Validate(); err != nil {
		return abi.Event{}, nil, err
	}
	if eventDef.Anonymous {
		return eventDef, topics, nil
	}
	if len(topics) == 0 {
		return abi.Event{}, nil, errNoEventSignature
	}
	if topics[0] != common.HashToLogTopic(eventDef.ID) {
		return abi.Event{}, nil, errEventSignatureMismatch
	}
	return eventDef, topics[1:], nil
}

// copyEventValues assigns decoded non-indexed event values to their generated
// struct fields. Event structs carry tags for every input (indexed and
// non-indexed), so the generic ABI tuple copier cannot be used on only the
// non-indexed subset without treating the other tags as errors.
func copyEventValues(out any, arguments abi.Arguments, indexed bool, values []any) error {
	destination := reflect.ValueOf(out)
	if !destination.IsValid() || destination.Kind() != reflect.Ptr || destination.IsNil() || destination.Elem().Kind() != reflect.Struct {
		return errors.New("abi: event output must be a non-nil pointer to a struct")
	}
	expected := 0
	for _, argument := range arguments {
		if argument.Indexed == indexed {
			expected++
		}
	}
	if expected != len(values) {
		return fmt.Errorf("abi: event argument/value count mismatch: %d arguments, %d values", expected, len(values))
	}
	destination = destination.Elem()
	fieldIndices, err := eventFieldIndices(destination.Type(), arguments)
	if err != nil {
		return err
	}
	valueIndex := 0
	for argumentIndex, argument := range arguments {
		if argument.Indexed != indexed {
			continue
		}
		if err := setEventValue(destination.Field(fieldIndices[argumentIndex]), values[valueIndex]); err != nil {
			return fmt.Errorf("abi: cannot unmarshal event field %q: %w", argument.Name, err)
		}
		valueIndex++
	}
	return nil
}

func eventFieldIndices(typ reflect.Type, arguments abi.Arguments) ([]int, error) {
	indices := make([]int, len(arguments))
	usedFields := make(map[int]bool)
	for i, argument := range arguments {
		fieldIndex, err := eventFieldIndex(typ, argument.Name, usedFields)
		if err != nil {
			return nil, err
		}
		usedFields[fieldIndex] = true
		indices[i] = fieldIndex
	}
	return indices, nil
}

// copyEventTopics decodes indexed values into a temporary struct and then
// copies them through the full event-input mapping. Decoding the indexed and
// non-indexed subsets directly into the destination would restart duplicate
// ABI-name occurrence counting for each subset and overwrite the first field.
func copyEventTopics(out any, arguments, indexed abi.Arguments, topics []common.LogTopic) error {
	fields := make([]reflect.StructField, len(indexed))
	for i, argument := range indexed {
		fieldType := argument.Type.GetType()
		switch argument.Type.T {
		case abi.StringTy, abi.BytesTy, abi.SliceTy, abi.ArrayTy, abi.TupleTy:
			fieldType = reflect.TypeFor[common.Hash]()
		}
		fields[i] = reflect.StructField{
			Name: fmt.Sprintf("Topic%d", i),
			Type: fieldType,
			Tag:  reflect.StructTag(fmt.Sprintf(`abi:%q`, argument.Name)),
		}
	}
	temporary := reflect.New(reflect.StructOf(fields))
	if err := abi.ParseTopics(temporary.Interface(), indexed, topics); err != nil {
		return err
	}
	values := make([]any, len(indexed))
	for i := range values {
		values[i] = temporary.Elem().Field(i).Interface()
	}
	return copyEventValues(out, arguments, true, values)
}

// eventFieldIndex prefers explicit ABI tags and falls back to the historical
// camel-case convention for handwritten event structs without tags. Tracking
// used fields also gives duplicate ABI names stable occurrence-based pairing.
func eventFieldIndex(typ reflect.Type, name string, used map[int]bool) (int, error) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" || used[i] {
			continue
		}
		if tag, ok := field.Tag.Lookup("abi"); ok && tag == name {
			return i, nil
		}
	}
	wanted := abi.ToCamelCase(name)
	if wanted == "" {
		return 0, errors.New("abi: purely underscored event field cannot unpack to struct")
	}
	if field, ok := typ.FieldByName(wanted); ok && len(field.Index) == 1 && !used[field.Index[0]] {
		if tag, tagged := field.Tag.Lookup("abi"); !tagged || tag == name {
			return field.Index[0], nil
		}
	}
	return 0, fmt.Errorf("abi: struct field for event argument %q not found", name)
}

// setEventValue delegates recursive tuple/slice/array conversion to the ABI
// package while turning its legacy panic-on-conversion API into a normal error.
func setEventValue(destination reflect.Value, value any) (err error) {
	if !destination.CanSet() {
		return fmt.Errorf("destination %s cannot be set", destination.Type())
	}
	if value == nil {
		destination.SetZero()
		return nil
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%v", recovered)
		}
	}()
	converted := reflect.ValueOf(abi.ConvertType(value, reflect.New(destination.Type()).Interface()))
	if converted.Kind() != reflect.Ptr || converted.IsNil() {
		return fmt.Errorf("conversion returned %s, want pointer to %s", converted.Type(), destination.Type())
	}
	destination.Set(converted.Elem())
	return nil
}

// ensureContext is a helper method to ensure a context is not nil, even if the
// user specified it as such.
func ensureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
