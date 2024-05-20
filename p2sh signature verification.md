Signature verification is tricky for p2sh transaction. Since it has several keys and signatures and order of signature should be the same as public key which means if the order of signature is n then its conrresponding
publick key should have order m >= n.

Now comes to how to construct the signature message z. Let's take it step by step, the following data chunk is a p2sh transaction and the chunk in { and } is the scriptsig of ints input:

```g
0100000001868278ed6ddfb6c1ed3ad5f8181eb0c7a385aa0836f01d5e4789e6bd304d87221a
000000
{
db00483045022100dc92655fe37036f47756db8102e0d7d5e28b3beb83a8fef4f5dc05
59bddfb94e02205a36d4e4e6c7fcd16658c50783e00c341609977aed3ad00937bf4ee942a89937
0148304502210Oda6bee3c93766232079a01639d07fa869598749729ae323eab8eef53577d611b02207bef
15429dcadce2121ea07f233115c6f09034c0be68db99980b9a6c5e75402201475221022626e955ea6ea6d9
8850c994f9107b036b1334f18ca8830bfff1295d21cfdb702103b287eaf122eea69030a0e9feed096bed8045c8b
98bec453e1ffac7fbdbd4bb7152ae
}
ffffffff04d3b11400000000001976a914904a49878c0adfc3aa05de7afad2c
c15f483a56a88ac7f400900000000001976a914418327e3f3dda4cf5b9089325a4b95abdfa0334088ac722c0c0000
0000001976a914ba35042cfe9fc66fd35ac2224eebdafd1028ad2788acdc4ace020000000017a91474d691da1574e
6b3c192ecfb52cc8984ee7b6c568700000000

```

1, find the scriptsig of input replace it with 00 as following:

```g
0100000001868278ed6ddfb6c1ed3ad5f8181eb0c7a385aa0836f01d5e4789e6bd304d87221a
000000
{
00
}
ffffffff04d3b11400000000001976a914904a49878c0adfc3aa05de7afad2c
c15f483a56a88ac7f400900000000001976a914418327e3f3dda4cf5b9089325a4b95abdfa0334088ac722c0c0000
0000001976a914ba35042cfe9fc66fd35ac2224eebdafd1028ad2788acdc4ace020000000017a91474d691da1574e
6b3c192ecfb52cc8984ee7b6c568700000000
```

2, the owner of p2sh transaction has the binary content of redeemscript, and we use the binary content of redeemscript to replace the 00 above, for example the content of given redeemscript is :
```g
475221022626e955ea6ea6d98850c994f9107b036b1334f18ca8830bfff1295d21cfdb702103b287eaf122eea69030a0e9feed096bed8045c8b98bec453e1ffac7fbdbd4bb7152ae
```
Then after replacing the 00 above we have following:
```g
0100000001868278ed6ddfb6c1ed3ad5f8181eb0c7a385aa0836f01d5e4789e6bd304d87221a0000000
{
475221022626e955ea6ea6d98850c994f9107b036b1334f18ca8830bfff1295d21cfdb702103b287eaf122eea69030a0e9feed096bed8045c8b98bec453e1ffac7fbdbd4bb7152ae
}
ffffffff04d3b11400000000001976a914904a49878c0adfc3aa05de7afad2c
c15f483a56a88ac7f400900000000001976a914418327e3f3dda4cf5b9089325a4b95abdfa0334088ac722c0c0000
0000001976a914ba35042cfe9fc66fd35ac2224eebdafd1028ad2788acdc4ace020000000017a91474d691da1574e
6b3c192ecfb52cc8984ee7b6c568700000000

```

3, convert the value of hash type SIGHASH_ALL into 4 bytes and append to the end of above data:
```g
0100000001868278ed6ddfb6c1ed3ad5f8181eb0c7a385aa0836f01d5e4789e6bd304d87221a000000
{
475221022626e955ea6ea6d98850c994f9107b036b1334f18ca8830bfff1295d21cfdb702103b287eaf122eea69030a0e9feed096bed8045c8b98bec453e1ffac7fbdbd4bb7152ae
}
ffffffff04d3b11400000000001976a914904a49878c0adfc3aa05de7afad2c
c15f483a56a88ac7f400900000000001976a914418327e3f3dda4cf5b9089325a4b95abdfa0334088ac722c0c0000
0000001976a914ba35042cfe9fc66fd35ac2224eebdafd1028ad2788acdc4ace020000000017a91474d691da1574e
6b3c192ecfb52cc8984ee7b6c568700000000[01000000]
```

4. Do hash256 on the above data and the result is the signature message.

Let's put the above steps into code in main.go:
```
import (
	ecc "elliptic_curve"
	"encoding/hex"
	"fmt"
	"math/big"
)

func main() {
	p2shRawData, err := hex.DecodeString("0100000001868278ed6ddfb6c1ed3ad5f8181eb0c7a385aa0836f01d5e4789e6bd304d87221a000000475221022626e955ea6ea6d98850c994f9107b036b1334f18ca8830bfff1295d21cfdb702103b287eaf122eea69030a0e9feed096bed8045c8b98bec453e1ffac7fbdbd4bb7152aeffffffff04d3b11400000000001976a914904a49878c0adfc3aa05de7afad2cc15f483a56a88ac7f400900000000001976a914418327e3f3dda4cf5b9089325a4b95abdfa0334088ac722c0c00000000001976a914ba35042cfe9fc66fd35ac2224eebdafd1028ad2788acdc4ace020000000017a91474d691da1574e6b3c192ecfb52cc8984ee7b6c56870000000001000000")
	if err != nil {
		panic(err)
	}
	s256 := ecc.Hash256(string(p2shRawData))
	zMsg := new(big.Int)
	zMsg.SetBytes(s256)
	fmt.Printf("signature message is :%x\n", zMsg)
}
```
Running the above code has the following result:
```g
signature message is :e71bfa115715d6fd33796948126f40a8cdd39f187e4afb03896795189fe1423c
```

Let's see how we can do the above steps in the process of verify, when we do SignHash on the given input, we need to check its conrresponding scriptPubKey is p2sh or not,
if it is, then we will use the redeemscript to replace the current scriptSig instead of using the ScriptPubKey, therefore we add code in input.go as following:
```g
func (t *TransactinInput) isP2sh(script *ScriptSig) bool {
	isP2sh := true
	if len(script.bitcoinOpCode.cmds[0]) != 1 || script.bitcoinOpCode.cmds[0][0] != OP_HASH160 {
		//the first element should be OP_HASH160
		isP2sh = false
	}

	if len(script.bitcoinOpCode.cmds[1]) == 1 {
		//the second element should be hash data chunk
		isP2sh = false
	}

	if len(script.bitcoinOpCode.cmds[2]) != 1 || script.bitcoinOpCode.cmds[2][0] != OP_EQUAL {
		//the third element should be OP_EQUAL
		isP2sh = false
	}

	return isP2sh
}

func (t *TransactinInput) ReplaceWithScriptPubKey(testnet bool) {
	//use scriptpubkey of previous transaction to replace current scriptsig
	tx := t.getPreviousTx(testnet)
	script := tx.txOutputs[t.previousTransactionIdex.Int64()].scriptPubKey
	isP2sh := t.isP2sh(script)
	if isP2sh == false {
		t.scriptSig = script
	} else {
		/*
			for p2sh we need to use redeemscript to replace the input script,
			the redeemscript is at the bottom of scriptSig command stack
		*/
		redeemscriptBinary := t.scriptSig.bitcoinOpCode.cmds[len(t.scriptSig.bitcoinOpCode.cmds)-1]
		redeemScriptReader := bytes.NewReader(redeemscriptBinary)
		redeemScript := NewScriptSig(bufio.NewReader(redeemScriptReader))
		t.scriptSig = redeemScript
	}
}
```

After completing above code, let's try to verify a p2sh transaction as following:
```g
func main() {
	p2shRawData, err := hex.DecodeString("0100000001868278ed6ddfb6c1ed3ad5f8181eb0c7a385aa0836f01d5e4789e6bd304d87221a000000db00483045022100dc92655fe37036f47756db8102e0d7d5e28b3beb83a8fef4f5dc0559bddfb94e02205a36d4e4e6c7fcd16658c50783e00c341609977aed3ad00937bf4ee942a8993701483045022100da6bee3c93766232079a01639d07fa869598749729ae323eab8eef53577d611b02207bef15429dcadce2121ea07f233115c6f09034c0be68db99980b9a6c5e75402201475221022626e955ea6ea6d98850c994f9107b036b1334f18ca8830bfff1295d21cfdb702103b287eaf122eea69030a0e9feed096bed8045c8b98bec453e1ffac7fbdbd4bb7152aeffffffff04d3b11400000000001976a914904a49878c0adfc3aa05de7afad2cc15f483a56a88ac7f400900000000001976a914418327e3f3dda4cf5b9089325a4b95abdfa0334088ac722c0c00000000001976a914ba35042cfe9fc66fd35ac2224eebdafd1028ad2788acdc4ace020000000017a91474d691da1574e6b3c192ecfb52cc8984ee7b6c568700000000")
	if err != nil {
		panic(err)
	}
	p2shTransaction := tx.ParseTransaction(p2shRawData)
	res := p2shTransaction.Verify()
	fmt.Printf("verify result of p2sh transaction is %v\n", res)
}
```
