package pivx

import (
	"encoding/json"

	"github.com/golang/glog"
	"github.com/trezor/blockbook/bchain"
	"github.com/trezor/blockbook/bchain/coins/btc"
)

// PivXRPC is an interface to JSON-RPC bitcoind service.
type PivXRPC struct {
	*btc.BitcoinRPC
    BitcoinGetChainInfo func() (*bchain.ChainInfo, error)
}

// NewPivXRPC returns new PivXRPC instance.
func NewPivXRPC(config json.RawMessage, pushHandler func(bchain.NotificationType)) (bchain.BlockChain, error) {
	b, err := btc.NewBitcoinRPC(config, pushHandler)
	if err != nil {
		return nil, err
	}

	s := &PivXRPC{
		b.(*btc.BitcoinRPC),
        b.GetChainInfo,
	}
	s.RPCMarshaler = btc.JSONMarshalerV1{}
	s.ChainConfig.SupportsEstimateFee = true
	s.ChainConfig.SupportsEstimateSmartFee = false

	return s, nil
}

// Initialize initializes PivXRPC instance.
func (b *PivXRPC) Initialize() error {
	ci, err := b.GetChainInfo()
	if err != nil {
		return err
	}
	chainName := ci.Chain

	glog.Info("Chain name ", chainName)
	params := GetChainParams(chainName)

	// always create parser
	b.Parser = NewPivXParser(params, b.ChainConfig)

	// parameters for getInfo request
	if params.Net == MainnetMagic {
		b.Testnet = false
		b.Network = "livenet"
	} else {
		b.Testnet = true
		b.Network = "testnet"
	}

	glog.Info("rpc: block chain ", params.Name)

	return nil
}


// getinfo

type CmdGetInfo struct {
	Method string `json:"method"`
}

type ResGetInfo struct {
	Error  *bchain.RPCError `json:"error"`
	Result struct {
        MoneySupply   json.Number `json:"moneysupply"`
        ZerocoinSupply  bchain.ZCdenoms    `json:"zPIVsupply"`
	} `json:"result"`
}

// getmasternodecount

type CmdGetMasternodeCount struct {
	Method string `json:"method"`
}

type ResGetMasternodeCount struct {
	Error  *bchain.RPCError `json:"error"`
    Result struct {
        Total   int    `json:"total"`
        Stable   int    `json:"stable"`
        Enabled   int    `json:"enabled"`
        InQueue   int    `json:"inqueue"`
	} `json:"result"`
}

// getnextsuperblock

type CmdGetNextSuperblock struct {
	Method string `json:"method"`
}

type ResGetNextSuperblock struct {
	Error  *bchain.RPCError `json:"error"`
	Result uint32           `json:"result"`
}

// GetChainInfo returns information about the connected backend
// PIVX adds MoneySupply and ZerocoinSupply to btc implementation
func (b *PivXRPC) GetChainInfo() (*bchain.ChainInfo, error) {
    rv, err := b.BitcoinGetChainInfo()
    if err != nil {
        return nil, err
    }

	glog.V(1).Info("rpc: getinfo")

    resGi := ResGetInfo{}
    err = b.Call(&CmdGetInfo{Method: "getinfo"}, &resGi)
    if err != nil {
        return nil, err
    }
    if resGi.Error != nil {
        return nil, resGi.Error
    }
    rv.MoneySupply = resGi.Result.MoneySupply
    rv.ZerocoinSupply = resGi.Result.ZerocoinSupply

    glog.V(1).Info("rpc: getmasternodecount")

    resMc := ResGetMasternodeCount{}
    err = b.Call(&CmdGetMasternodeCount{Method: "getmasternodecount"}, &resMc)
    if err != nil {
        return nil, err
    }
    if resMc.Error != nil {
        return nil, resMc.Error
    }
    rv.MasternodeCount = resMc.Result.Enabled

    glog.V(1).Info("rpc: getnextsuperblock")

    resNs := ResGetNextSuperblock{}
    err = b.Call(&CmdGetNextSuperblock{Method: "getnextsuperblock"}, &resNs)
    if err != nil {
        return nil, err
    }
    if resNs.Error != nil {
        return nil, resNs.Error
    }
    rv.NextSuperBlock = resNs.Result

	return rv, nil
}
