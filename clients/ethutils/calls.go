package ethutils

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/vegaprotocol/vega-monitoring/config"
)

type EthCall struct {
	id                  string
	abi                 string
	abiObject           abi.ABI
	address             common.Address
	methodName          string
	args                []interface{}
	call                []byte
	outputIndex         int
	outputTransformFunc func(interface{}) interface{}
}

func NewEthCallFromConfig(cfg config.EthCall) (*EthCall, error) {
	return NewEthCall(
		cfg.Name,
		cfg.ABI,
		cfg.Address,
		cfg.Method,
		cfg.Args,
		cfg.OutputIndex,
		cfg.OutputTransform,
	)
}

func NewEthCall(
	id string,
	abiDefinition string,
	address string,
	methodName string,
	args []interface{},
	outputIndex int,
	transformOutput string,
) (*EthCall, error) {
	packableArgs, err := constructPackableArgs(abiDefinition, methodName, args)
	if err != nil {
		return nil, fmt.Errorf("failed to construct packable args: %w", err)
	}

	abiObject, err := abi.JSON(strings.NewReader(abiDefinition))
	if err != nil {
		return nil, fmt.Errorf("failed to read abi json to object: %w", err)
	}

	callBytes, err := abiObject.Pack(methodName, packableArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack arguments to the raw transaction: %w", err)
	}

	transformFunc, err := getTransformFunction(transformOutput)
	if err != nil {
		return nil, fmt.Errorf("failed create output transform function: %w", err)
	}

	return &EthCall{
		id:                  id,
		abi:                 abiDefinition,
		abiObject:           abiObject,
		address:             common.HexToAddress(address),
		methodName:          methodName,
		args:                packableArgs,
		call:                callBytes,
		outputIndex:         outputIndex,
		outputTransformFunc: transformFunc,
	}, nil
}

func getTransformFunction(transformFormula string) (func(interface{}) interface{}, error) {
	if strings.HasPrefix(transformFormula, "float_price:") {
		decimalPlacesString := strings.TrimPrefix(transformFormula, "float_price:")

		decimalPlaces, err := strconv.Atoi(decimalPlacesString)
		if err != nil {
			return nil, fmt.Errorf("invalid arguments for the `float_price` transformation function: function accepts only one argument and it should be decimal places for the asset")
		}

		return func(input interface{}) interface{} {
			bigIntInput, ok := input.(*big.Int)
			if !ok {
				return 0.0
			}

			// Scale to float64
			powInt := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimalPlaces)), nil)
			result, _ := new(big.Float).Quo(new(big.Float).SetInt(bigIntInput), new(big.Float).SetInt(powInt)).Float64()

			return result
		}, nil
	}

	// default function no transformation
	return func(input interface{}) interface{} {
		return input
	}, nil
}

func (ec *EthCall) Call(ctx context.Context, client *ethclient.Client) (interface{}, error) {
	msg := ethereum.CallMsg{
		To:   &ec.address,
		Data: ec.call,
	}

	outputBytes, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	outputs, err := ec.abiObject.Methods[ec.methodName].Outputs.Unpack(outputBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack contract call output: %w", err)
	}

	finalOutput := outputs[ec.outputIndex]

	return ec.outputTransformFunc(finalOutput), nil
}

func constructPackableArgs(abiDefinition string, methodName string, args []interface{}) ([]interface{}, error) {
	abi, err := abi.JSON(strings.NewReader(abiDefinition))
	if err != nil {
		return nil, fmt.Errorf("failed to create ABI from the JSON: %w", err)
	}

	method, methodExists := abi.Methods[methodName]
	if !methodExists {
		return nil, fmt.Errorf("the %s method does not exist in the given abi", method)
	}

	if len(args) != len(method.Inputs) {
		return nil, fmt.Errorf(
			"invalid arguments passed: abi contains %d inputs, %d args passed",
			len(method.Inputs),
			len(args),
		)
	}

	result := []interface{}{}
	for idx, arg := range args {
		packableArg, err := packableValue(arg, method.Inputs[idx])
		if err != nil {
			return nil, fmt.Errorf("failed to prepare packable value for %v: %w", arg, err)
		}

		result = append(result, packableArg)
	}

	return result, nil
}

// https://github.com/ethereum/go-ethereum/blob/master/accounts/abi/type_test.go
func packableValue(arg interface{}, abiArg abi.Argument) (interface{}, error) {
	switch v := arg.(type) {
	case string:
		if strings.HasPrefix(v, "int:") {
			vInt := strings.TrimPrefix(v, "int:")
			if abiArg.Type.T != abi.IntTy && abiArg.Type.T != abi.UintTy {
				return nil, fmt.Errorf("type mismatch, bigint(%s) passed, %#v expected", vInt, abiArg.Type.T)
			}

			bigInt, success := big.NewInt(0).SetString(vInt, 10)

			if !success {
				return nil, fmt.Errorf("failed to create bigint from given string: %s", vInt)
			}
			return bigInt, nil
		}

		if abiArg.Type.T == abi.StringTy {
			return v, nil
		}

		vTrimmed := strings.TrimPrefix(v, "0x")
		decodedString, err := hex.DecodeString(vTrimmed)
		if err != nil {
			return nil, fmt.Errorf("failed to decode string(%s) into hex: %w", v, err)
		}

		switch abiArg.Type.T {
		case abi.FixedBytesTy:
			var result [32]byte

			copy(result[:], decodedString)
			return result, nil
		case abi.AddressTy:
			return common.HexToAddress(v), nil
		default:
			return nil, fmt.Errorf("unsupported type(%#v) for the %s string", abiArg.Type.T, v)
		}
	case int:
		if abiArg.Type.T != abi.IntTy {
			return nil, fmt.Errorf("type mismatch, int32 value passed, abi expect type %#v", abiArg.Type.T)
		}
		if abiArg.Type.Size != 32 {
			return nil, fmt.Errorf("size mismatch, int32 passed, int%d expected", abiArg.Type.Size)
		}

		return int32(v), nil
	case int32:
		if abiArg.Type.T != abi.IntTy {
			return nil, fmt.Errorf("type mismatch, int32 value passed, abi expect type %#v", abiArg.Type.T)
		}
		if abiArg.Type.Size != 32 {
			return nil, fmt.Errorf("size mismatch, int32 passed, int%d expected", abiArg.Type.Size)
		}

		return int32(v), nil
	case int16:
		if abiArg.Type.T != abi.IntTy {
			return nil, fmt.Errorf("type mismatch, int16 value passed, abi expect type %#v", abiArg.Type.T)
		}
		if abiArg.Type.Size != 16 {
			return nil, fmt.Errorf("size mismatch, int16 passed, int%d expected", abiArg.Type.Size)
		}

		return int16(v), nil
	case int64:
		if abiArg.Type.T != abi.IntTy {
			return nil, fmt.Errorf("type mismatch, int64 value passed, abi expect type %#v", abiArg.Type.T)
		}
		if abiArg.Type.Size != 64 {
			return nil, fmt.Errorf("size mismatch, int64 passed, int%d expected", abiArg.Type.Size)
		}

		return int64(v), nil

	case uint:
		if abiArg.Type.T != abi.UintTy {
			return nil, fmt.Errorf("type mismatch, uint32 value passed, abi expect type %#v", abiArg.Type.T)
		}
		if abiArg.Type.Size != 32 {
			return nil, fmt.Errorf("size mismatch, uint32 passed, int%d expected", abiArg.Type.Size)
		}

		return uint32(v), nil
	case uint32:
		if abiArg.Type.T != abi.UintTy {
			return nil, fmt.Errorf("type mismatch, uint32 value passed, abi expect type %#v", abiArg.Type.T)
		}
		if abiArg.Type.Size != 32 {
			return nil, fmt.Errorf("size mismatch, uint32 passed, int%d expected", abiArg.Type.Size)
		}

		return uint32(v), nil
	case uint16:
		if abiArg.Type.T != abi.UintTy {
			return nil, fmt.Errorf("type mismatch, uint16 value passed, abi expect type %#v", abiArg.Type.T)
		}
		if abiArg.Type.Size != 16 {
			return nil, fmt.Errorf("size mismatch, uint16 passed, uint%d expected", abiArg.Type.Size)
		}

		return uint16(v), nil
	case uint64:
		if abiArg.Type.T != abi.UintTy {
			return nil, fmt.Errorf("type mismatch, uint64 value passed, abi expect type %#v", abiArg.Type.T)
		}
		if abiArg.Type.Size != 64 {
			return nil, fmt.Errorf("size mismatch, uint64 passed, uint%d expected", abiArg.Type.Size)
		}

		return uint64(v), nil

	}

	return nil, fmt.Errorf("unsupported type")
}

func (ec EthCall) ContractAddress() common.Address {
	return ec.address
}

func (ec EthCall) MethodName() string {
	return ec.methodName
}

func (ec EthCall) ID() string {
	return ec.id
}
