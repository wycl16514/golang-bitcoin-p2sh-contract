package main

import (
	"encoding/hex"
	"fmt"
	tx "transaction"
)

/*
1. make sure the total amount in the inputs of transaction is more than
than ouput,
*/

func main() {
	p2shRawData, err := hex.DecodeString("0100000001868278ed6ddfb6c1ed3ad5f8181eb0c7a385aa0836f01d5e4789e6bd304d87221a000000db00483045022100dc92655fe37036f47756db8102e0d7d5e28b3beb83a8fef4f5dc0559bddfb94e02205a36d4e4e6c7fcd16658c50783e00c341609977aed3ad00937bf4ee942a8993701483045022100da6bee3c93766232079a01639d07fa869598749729ae323eab8eef53577d611b02207bef15429dcadce2121ea07f233115c6f09034c0be68db99980b9a6c5e75402201475221022626e955ea6ea6d98850c994f9107b036b1334f18ca8830bfff1295d21cfdb702103b287eaf122eea69030a0e9feed096bed8045c8b98bec453e1ffac7fbdbd4bb7152aeffffffff04d3b11400000000001976a914904a49878c0adfc3aa05de7afad2cc15f483a56a88ac7f400900000000001976a914418327e3f3dda4cf5b9089325a4b95abdfa0334088ac722c0c00000000001976a914ba35042cfe9fc66fd35ac2224eebdafd1028ad2788acdc4ace020000000017a91474d691da1574e6b3c192ecfb52cc8984ee7b6c568700000000")
	if err != nil {
		panic(err)
	}
	p2shTransaction := tx.ParseTransaction(p2shRawData)
	fmt.Printf("p2sh transaction details: %s\n", p2shTransaction)
	res := p2shTransaction.Verify()
	fmt.Printf("verify result of p2sh transaction is %v\n", res)
}

/*
1. find the scriptsig for the current input

2. replace the scriptsig data with 00

3. use the scriptpubkey from previous transaction to replace the 00

4. append hash type to the end of the transaction binary data
hash type is 4 byte in little endian format

SIGHASH_ALL 1 => 01 00 00 00

5. Do hash256 on the modified binary data

=> signature message

0100000001813f79011acb80925dfe69b3def355fe914bd1d96a3f5f71bf8303c6a989c7d100000000


	1976a914a802fc56c704ce87c42d7c92eb75e7896bdc41ae88ac


feffffff02a135ef01000000001976a914bc3b654dca7e56b04dca18f2566cdaf02e8d9ada88
ac99c39800000000001976a9141c4bc762dd5423e332166702cb75f40df79fea1288ac1943060001000000
*/
