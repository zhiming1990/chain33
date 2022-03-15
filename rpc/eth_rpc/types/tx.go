package types

import (
	"errors"
	"fmt"
	ctypes "github.com/33cn/chain33/types"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
)

func DecodeSignature(sig []byte) (r, s, v *big.Int) {
	if len(sig) != crypto.SignatureLength {
		panic(fmt.Sprintf("wrong size for signature: got %d, want %d", len(sig), crypto.SignatureLength))
	}
	r = new(big.Int).SetBytes(sig[:32])
	s = new(big.Int).SetBytes(sig[32:64])
	v = new(big.Int).SetBytes([]byte{sig[64] + 27})
	return r, s, v
}
func MakeDERSigToRSV(eipSigner etypes.EIP155Signer,sig []byte)(r,s,v *big.Int,err error){
	rb,sb,err := paraseDERCode(sig)
	if err!=nil{
		return
	}
	var signature []byte
	signature=append(signature,rb...)
	signature=append(signature,sb...)
	signature=append(signature,0x00)
	r,s,v =  DecodeSignature(signature)
	if eipSigner.ChainID().Sign() != 0 {
		v = big.NewInt(int64(signature[64] + 35))
		v.Add(v, new(big.Int).Mul(eipSigner.ChainID(), big.NewInt(2)))
	}
	return
}

func  paraseDERCode(sig []byte)(r,s []byte,err error){
	if len(sig)!=70{
		return nil,nil,errors.New("wrong sig data")
	}
	if sig[0]==0x30&&sig[3]==0x20{
		r=sig[4:36]
	}
	if sig[37]==0x20{
		s=sig[38:70]
	}

	return
}

/*
Tx         *Transaction `protobuf:"bytes,1,opt,name=tx,proto3" json:"tx,omitempty"`
	Receipt    *ReceiptData `protobuf:"bytes,2,opt,name=receipt,proto3" json:"receipt,omitempty"`
	Proofs     [][]byte     `protobuf:"bytes,3,rep,name=proofs,proto3" json:"proofs,omitempty"`
	Height     int64        `protobuf:"varint,4,opt,name=height,proto3" json:"height,omitempty"`
	Index      int64        `protobuf:"varint,5,opt,name=index,proto3" json:"index,omitempty"`
	Blocktime  int64        `protobuf:"varint,6,opt,name=blocktime,proto3" json:"blocktime,omitempty"`
	Amount     int64        `protobuf:"varint,7,opt,name=amount,proto3" json:"amount,omitempty"`
	Fromaddr   string       `protobuf:"bytes,8,opt,name=fromaddr,proto3" json:"fromaddr,omitempty"`
	ActionName string       `protobuf:"bytes,9,opt,name=actionName,proto3" json:"actionName,omitempty"`
	Assets     []*Asset     `protobuf:"bytes,10,rep,name=assets,proto3" json:"assets,omitempty"`
	TxProofs   []*TxProof   `protobuf:"bytes,11,rep,name=txProofs,proto3" json:"txProofs,omitempty"`
	FullHash   []byte       `protobuf:"bytes,12,opt,name=fullHash,proto3" json:"fullHash,omitempty"`

*/
func TxDetailsToEthTx(txdetails *ctypes.TransactionDetails,cfg *ctypes.Chain33Config )(txs Transactions, err error){
	for _,detail:=range txdetails.GetTxs(){
		var tx Transaction
		tx.To= detail.Tx.To
		tx.From=detail.Fromaddr
		tx.Value= fmt.Sprintf("0x%x",detail.GetAmount())//big.NewInt(detail.GetAmount()).
		tx.Type=fmt.Sprint(detail.Receipt.Ty)
		tx.BlockNumber=fmt.Sprintf("0x%x",detail.Height)
		tx.TransactionIndex=fmt.Sprintf("0x%x",detail.GetIndex())
		eipSigner:= etypes.NewEIP155Signer(big.NewInt(int64(cfg.GetChainID())))
		r,s,v,err:=MakeDERSigToRSV(eipSigner,detail.Tx.GetSignature().GetSignature())
		if err!=nil{
			return nil,err
		}
		tx.V=v
		tx.R=r
		tx.S=s
		txs=append(txs,&tx)
	}
	return

}