package jail

import (
	"context"
	"encoding/json"
	"math/big"
	"time"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/robertkrimen/otto"
	"github.com/status-im/status-go/geth/common"
)

// ExecutionPolicy provides a central container for the executions of RPCCall requests for both
// remote/upstream processing and internal node processing.
type ExecutionPolicy struct{}

// ExecuteSendTransaction defines a function to execute RPC requests for eth_sendTransaction method only.
func (ep ExecutionPolicy) ExecuteSendTransaction(manager common.NodeManager, account common.AccountManager, req common.RPCCall, call otto.FunctionCall) (*otto.Object, error) {
	config, err := manager.NodeConfig()
	if err != nil {
		return nil, err
	}

	if config.UpstreamConfig.Enabled {
		return ep.ExecuteRemoteSendTransaction(manager, account, req, call)
	}

	return ep.ExecuteLocalSendTransaction(manager, req, call)
}

// ExecuteRemoteSendTransaction defines a function to execute RPC method eth_sendTransaction over the upstream server.
func (ExecutionPolicy) ExecuteRemoteSendTransaction(manager common.NodeManager, account common.AccountManager, req common.RPCCall, call otto.FunctionCall) (*otto.Object, error) {
	config, err := manager.NodeConfig()
	if err != nil {
		return nil, err
	}

	selectedAcct, err := account.SelectedAccount()
	if err != nil {
		return nil, err
	}

	client, err := manager.RPCClient()
	if err != nil {
		return nil, err
	}

	fromAddr, err := req.ParseFromAddress()
	if err != nil {
		return nil, err
	}

	toAddr, err := req.ParseToAddress()
	if err != nil {
		return nil, err
	}

	// We need to request a new transaction nounce from upstream node.
	ctx, canceller := context.WithDeadline(context.Background(), time.Now().Add(1*time.Minute))
	defer canceller()

	var num hexutil.Uint
	if err := client.CallContext(ctx, &num, "eth_getTransactionCount", fromAddr, "latest"); err != nil {
		return nil, err
	}

	nonce := uint64(num)
	gas := (*big.Int)(req.ParseGas())
	dataVal := []byte(req.ParseData())
	priceVal := (*big.Int)(req.ParseValue())
	gasPrice := (*big.Int)(req.ParseGasPrice())
	chainID := big.NewInt(int64(config.NetworkID))

	tx := types.NewTransaction(nonce, toAddr, priceVal, gas, gasPrice, dataVal)
	txs, err := types.SignTx(tx, types.NewEIP155Signer(chainID), selectedAcct.AccountKey.PrivateKey)
	if err != nil {
		return nil, err
	}

	// Attempt to get the hex version of the transaction.
	txBytes, err := rlp.EncodeToBytes(txs)
	if err != nil {
		return nil, err
	}

	ctx2, canceler2 := context.WithDeadline(context.Background(), time.Now().Add(1*time.Minute))
	defer canceler2()

	var result json.RawMessage
	if err := client.CallContext(ctx2, &result, "eth_sendRawTransaction", gethcommon.ToHex(txBytes)); err != nil {
		return nil, err
	}

	resp, err := call.Otto.Object(`({"jsonrpc":"2.0"})`)
	if err != nil {
		return nil, err
	}

	resp.Set("id", req.ID)
	resp.Set("result", result)
	resp.Set("hash", txs.Hash().String())

	return resp, nil
}

// ExecuteLocalSendTransaction defines a function which handles execution of RPC method over the internal rpc server
// from the eth.LightClient. It specifically caters to process eth_sendTransaction.
func (ExecutionPolicy) ExecuteLocalSendTransaction(manager common.NodeManager, req common.RPCCall, call otto.FunctionCall) (*otto.Object, error) {
	resp, err := call.Otto.Object(`({"jsonrpc":"2.0"})`)
	if err != nil {
		return nil, err
	}

	resp.Set("id", req.ID)

	txHash, err := processRPCCall(manager, req, call)
	resp.Set("result", txHash.Hex())

	if err != nil {
		resp = newErrorResponse(call, -32603, err.Error(), &req.ID).Object()
		return resp, nil
	}

	return resp, nil
}

// ExecuteOtherTransaction defines a function which handles the processing of non `eth_sendTransaction`
// rpc request to the internal node server.
func (ExecutionPolicy) ExecuteOtherTransaction(manager common.NodeManager, req common.RPCCall, call otto.FunctionCall) (*otto.Object, error) {
	client, err := manager.RPCClient()
	if err != nil {
		return nil, common.StopRPCCallError{Err: err}
	}

	JSON, err := call.Otto.Object("JSON")
	if err != nil {
		return nil, err
	}

	var result json.RawMessage

	resp, _ := call.Otto.Object(`({"jsonrpc":"2.0"})`)
	resp.Set("id", req.ID)

	// do extra request pre processing (persist message id)
	// within function semaphore will be acquired and released,
	// so that no more than one client (per cell) can enter
	messageID, err := preProcessRequest(call.Otto, req)
	if err != nil {
		return nil, common.StopRPCCallError{Err: err}
	}

	err = client.Call(&result, req.Method, req.Params...)

	switch err := err.(type) {
	case nil:
		if result == nil {

			// Special case null because it is decoded as an empty
			// raw message for some reason.
			resp.Set("result", otto.NullValue())

		} else {

			resultVal, callErr := JSON.Call("parse", string(result))

			if callErr != nil {
				resp = newErrorResponse(call, -32603, callErr.Error(), &req.ID).Object()
			} else {
				resp.Set("result", resultVal)
			}

		}

	case rpc.Error:

		resp.Set("error", map[string]interface{}{
			"code":    err.ErrorCode(),
			"message": err.Error(),
		})

	default:

		resp = newErrorResponse(call, -32603, err.Error(), &req.ID).Object()
	}

	// do extra request post processing (setting back tx context)
	postProcessRequest(call.Otto, req, messageID)

	return resp, nil
}
