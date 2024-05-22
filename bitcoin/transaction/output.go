package transaction

import (
	"bufio"
	"fmt"
	"io"
	"math/big"
)

type TransactionOutput struct {
	//satoshi
	amount       *big.Int
	scriptPubKey *ScriptSig
}

func InitTransactionOutput(amount *big.Int, script *ScriptSig) *TransactionOutput {
	return &TransactionOutput{
		amount:       amount,
		scriptPubKey: script,
	}
}

func (t *TransactionOutput) String() string {
	return fmt.Sprintf("amount:%v\n scriptPubKey: %x\n", t.amount, t.scriptPubKey.Serialize())
}

func NewTractionOutput(reader *bufio.Reader) *TransactionOutput {
	/*
		amount is in stashi 1/100,000,0000 of one bitcoin
	*/
	amountBuf := make([]byte, 8)
	//bug fix
	//reader.Read(amountBuf)
	io.ReadFull(reader, amountBuf)
	amount := LittleEndianToBigInt(amountBuf, LITTLE_ENDIAN_8_BYTES)
	script := NewScriptSig(reader)
	return &TransactionOutput{
		amount:       amount,
		scriptPubKey: script,
	}
}

func (t *TransactionOutput) Serialize() []byte {
	result := make([]byte, 0)
	result = append(result, BigIntToLittleEndian(t.amount, LITTLE_ENDIAN_8_BYTES)...)
	result = append(result, t.scriptPubKey.Serialize()...)
	return result
}
