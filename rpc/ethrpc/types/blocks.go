package types

import (
	"github.com/33cn/chain33/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func BlockDetailToEthBlock(details *types.BlockDetails,cfg *types.Chain33Config ,full bool )(*Block,error){
	var block Block
	var header Header
	fullblock:=details.GetItems()[0]
	header.Time= hexutil.Uint(fullblock.GetBlock().GetBlockTime()).String()
	header.Number=hexutil.Uint(fullblock.GetBlock().Height).String()
	header.TxHash=common.BytesToHash(fullblock.GetBlock().GetHeader(cfg).TxHash).Hex()
	header.Difficulty=hexutil.Uint(int64(fullblock.GetBlock().GetDifficulty())).String()
	header.ParentHash=common.BytesToHash(fullblock.GetBlock().ParentHash).Hex()
	header.Root=common.BytesToHash(fullblock.GetBlock().GetStateHash()).Hex()
	header.Coinbase=fullblock.GetBlock().GetTxs()[0].From()
	//暂不支持ReceiptHash,UncleHash
	//header.ReceiptHash=
	//header.UncleHash

	//处理交易
	txs,err:=TxsToEthTxs(fullblock.GetBlock().GetTxs(),cfg,full)
	if err!=nil{
		return nil,err
	}
	block.Header=&header
	block.Transactions=txs
	block.Hash=common.BytesToHash(fullblock.GetBlock().Hash(cfg)).Hex()
	return &block,nil
}

