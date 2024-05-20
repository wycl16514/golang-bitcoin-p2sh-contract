At the beginning of this section, we need to do some bug fixing. The first bug is when we use reader from bufio to read data into given buffer, we need to change it to:

io.ReadFull(reader, buffer)

The problem here is, we want the reader to fill data at the full of given buffer, but for the read method of reader from bufio, it may not fill the buffer to full at 
some cases, and we need to use io.ReadFull to make sure each time we fill the buffer at full.

The second fix is as following:
```g
func (s *ScriptSig) Evaluate(z []byte) bool {
    ...
    //bug fix, need to check the top of the stack
	if len(s.bitcoinOpCode.stack[len(s.bitcoinOpCode.stack)-1]) == 0 {
		return false
	}
    return true
}
```

For some situation, we may need multiple parties to control the release of fund. For example, if the fund need to proved by several board members, each one has his/her own private key, the fund can only be released
if all board members sign to the contract.

In order to support multiple private keys for one transaction, we need to use the op code with name OP_CHECKMULTISIG, its hex value is 0xae. This is a quit complicated opeation, it needs many elements on the stack
to operate, the following binary data is a contract with two public keys:

```g
514104fcf07bb1222f7925f2b7cc15183a40443c578e62ea17100aa
3b44ba66905c95d4980aec4cd2f6eb426d1b1ec45d76724f26901099416b9265b
76ba67c8b0b73d210202be80a0ca69c0e000b97d507f45b98c49f58fec6650b64ff70e6ffccc3e6d0052ae
```
Let's break it down into pieces:

1, the first byte 0x51 is an op code OP_1, which means push value 1 on to the stack

2, the second byte 0x41 is the length for the following data chunk which is the first public key

3, the following 0x41 bytes of data is the raw data of the first publick key:

04fcf07bb1222f7925f2b7cc15183a40443c578e62ea17100aa
3b44ba66905c95d4980aec4cd2f6eb426d1b1ec45d76724f26901099416b9265b
76ba67c8b0b73d

4. the byte following the end of first publick key is 0x21, its the length of the second public key

5. the 0x21 bytes following is the second public key:
0202be80a0ca69c0e000b97d507f45b98c49f58fec6650b64ff70e6ffccc3e6d00

6. The byte following the end of second public key is 0x52, it is an op code OP_2, which means push value 2 on the stack

7. The final byte is an op code OP_CHECKMULTISIG

The aboved script is scriptPubKey from the output of previous transaction. The conrresponding scriptSig in current transaction input is as following:
```g
00483045022100e222a0a6816475d85ad28fbeb66e97c
931081076dc9655da3afc6c1d81b43f9802204681f9ea9d52a31c
9c47cf78b71410ecae6188d7c31495f5f1adfeOdf5864a7401
```

1, the first byte 0x00 is op code OP_0, it is used to push an empty array on the stack

2, the second byte 0x48 is the length of signature

3, the data chunk following the second byte is belongs to signature:

3045022100e222a0a6816475d85ad28fbeb66e97c
931081076dc9655da3afc6c1d81b43f9802204681f9ea9d52a31c
9c47cf78b71410ecae6188d7c31495f5f1adfeOdf5864a7401

When we combine to ScriptPubKey and ScriptSig together as evaluation script, it will look like following:

![bitcoin_script](https://github.com/wycl16514/golang-bitcoin-p2sh-contract/assets/7506958/05d442e7-606b-405b-b1c0-22e88aac45fa)

Let's see the process of running the script above, first is OP_0 and following a signature data chunk, therefore there are directly push to the evaluate stack:

![bitcoin_script (1)](https://github.com/wycl16514/golang-bitcoin-p2sh-contract/assets/7506958/2aeac5d6-8f9b-4a38-a469-acfb67aa955c)

The OP_1 will push value 1 on to the stack and pubkey1 , pubkey2 are data elements and OP_2 is push value 2 on the stack as following:

![bitcoin_script (2)](https://github.com/wycl16514/golang-bitcoin-p2sh-contract/assets/7506958/af84209b-b8cb-42ec-b13d-bbd36c8dd6b2)

Now the script has only a OP_CHECKMULTISIG command, when executing this command, the first element is 2, then the virtual machine pop the top element from the stack and take the following 2 elements as public keys.
Then the top element of the evaluate stack is 1 which tells virtual machine to take the following one element as signature, if the two public keys can both verify the signature then the virtual machine will push an
element of 1 on to the stack to indicate the success of evaluation.

Now the question is why there is still an element 0 which is an empty slice on the stack? This is due to the design defect of bitcoin。Theoretically if there are m public keys and n signatures, the command needs
2 + m + n elements on the stack , but due to the design bug, it nees to take 3 + m +n elments, we need to push another empty element to overcome the defect. For most of bitcoin full nodes, when they are executing
the command OP_MULTICHECKSIG, if there is not OP_0 on the evaluate stack, they will just deem the script is fail.

The number of public key and signature may not constrain to 2 and 1, it is possible that there are m public keys and n signature, we need to check the value of top element to decide the number of public keys and
after reading the given number of public keys, we need to read the top element again to decide the number of signature, let's see how we can implement the OP_CHECKMULTISIG command, in op.go we have following code:

```g
func (b *BitcoinOpCode) opCheckMultiSig(zBin []byte) bool {
	if len(b.stack) < 1 {
		return false
	}
	//read the top element to get the number of public keys
	pubKeyCounts := int(b.DecodeNum(b.popStack()))
	if len(b.stack) < pubKeyCounts+1 {
		return false
	}

	secPubKeys := make([][]byte, 0)
	for i := 0; i < pubKeyCounts; i++ {
		secPubKeys = append(secPubKeys, b.popStack())
	}

	//get the number of signatures
	sigCounts := int(b.DecodeNum(b.popStack()))
	if len(b.stack) < sigCounts+1 {
		return false
	}

	derSignatures := make([][]byte, 0)
	for i := 0; i < sigCounts; i++ {
		signature := b.popStack()
		//remove last byte, it is hash type
		signature = signature[0 : len(signature)-1]
		derSignatures = append(derSignatures, signature)
	}

	points := make([]*ecc.Point, 0)
	sigs := make([]*ecc.Signature, 0)
	for i := 0; i < pubKeyCounts; i++ {
		points = append(points, ecc.ParseSEC(secPubKeys[i]))
	}
	for i := 0; i < sigCounts; i++ {
		sigs = append(sigs, ecc.ParseSigBin(derSignatures[i]))
	}
	/*
		for m public keys and n signatures, we need to make sure, there are
		n public keys from the total of m that can verify the n siganture
	*/
	z := new(big.Int)
	z.SetBytes(zBin)
	n := ecc.GetBitcoinValueN()
	zField := ecc.NewFieldElement(n, z)

	for _, sig := range sigs {
		if len(points) == 0 {
			return false
		}
		for len(points) > 0 {
			point := points[0]
			points = points[1:]
			if point.Verify(zField, sig) {
				break
			}
		}
	}
	b.stack = append(b.stack, b.EncodeNum(1))
	return true
}
```

As you can see the procedure above is quit ugly but it works. But the method above has several problems:

1, it nees many public keys and that will make the ScriptPubKey very long, which add to the cost of communication bandwidth.

2, It will takes out huge disk volumn and RAM for bitcoin full nodes.

3, It can be turned into attack method to harm the bitcoin blockchain.

In order to avoid the shortcomings, bitcoin community design the pay-to-script-hash(p2sh) script. When something is too long to handle, what kind of solution we can take? one method
that often used to handle such situation is hash, we can hash a very long string into a fixed length hash string .P2sh transaction is doing something
like this, for example look at the following multisig transaction:
```g
5221022626e955ea6ea6d98850c994f9107b036b1334f18ca8830bfff1295d21
cfdb702103b287eaf122eea69030a0e9feed096bed8045c8b98bec453e1ffac7fbdbd4bb7152ae
```
it is a ScriptPubKey with two publick keys , its structure is the same as the example above, the ScriptPubKey that contains multiple public key, has another name called redeemscript,
we do a Hash160 on the content above and get the following result:
```g
74d691da1574e6b3c192ecfb52cc8984ee7b6c56
```
Then we construct a script by using the data above, its raw data is as following:
```
a91474d691da1574e6b3c192ecfb52cc8984ee7b6c5687
```
The first byte a9 is command OP_HASH160

The second byte 0x14 is the length of the following hash data chunk

The following data chunk is the hash160 result of the above ScriptPubKey,

The last byte 0x87 is op code OP_EQUAL

Now, the original scriptPubKey will not contains in the output of previous transaction any more, the creator of p2sh transaction is responsible for saving the orignal ScriptPubKey 
which is the redeemscript, now the scriptpubkey and scriptsig looks like following:

![bitcoin_script (3)](https://github.com/wycl16514/golang-bitcoin-p2sh-contract/assets/7506958/9d9febb4-7a72-4681-9f25-771b8128327e)


Let's go through the execution process of aboved combined script, The first element is OP_0 then the script push an empty array onto the parsing stack and the following 3 elements are data elements and they will
push to the parsing stack directly:

![bitcoin_script (4)](https://github.com/wycl16514/golang-bitcoin-p2sh-contract/assets/7506958/d88af736-8b25-4cc6-890c-06cf66f4d457)

Now the top element of script is OP_HASH160, then the script will get the top element from the parsing script and do Hash160 computation on it and push the hash result to the top of parsing stack:


![bitcoin_script (5)](https://github.com/wycl16514/golang-bitcoin-p2sh-contract/assets/7506958/1db3202a-243d-4993-9369-cc42eacd99b0)

Now the top element of the script is a chunk of hash data, then the data will push to the top of parsing stack directly:

![bitcoin_script](https://github.com/wycl16514/golang-bitcoin-p2sh-contract/assets/7506958/296302a5-fee1-4c82-8ad3-1f807679ea07)

Now the top element on the script is OP_EQUAL, this command will take the top 2 elements of the parsing stack, and check if they are equal, if they are, the script will push a value 1 on the parsing stack:

![bitcoin_script (1)](https://github.com/wycl16514/golang-bitcoin-p2sh-contract/assets/7506958/9b44f736-a90c-4366-b4f9-c6b99572b86a)

If the bitcoin node is design before the release of protocol BIP0016, they will aggree to release the fund in the transaction, of couse now, most of bitcoin node are designed after protocol BIP0016, when they take
the first element from the parsing stack and found it is 1, then they will take out the redeemscript again and execute it, let's take a look on the script again:

```g
52 21

022626e955ea6ea6d98850c994f9107b036b1334f18ca8830bfff1295d21cfdb70

21

03b287eaf122eea69030a0e9feed096bed8045c8b98bec453e1ffac7fbdbd4bb71

52

ae
```
1, The first byte is 0x52, it is value for op code OP_2
2, The second byte is 0x21, is the length for the following data chunk which is the first public key
3, following the end of first public key is byte 0x21 again, which is length for the second public key
4, following the end of second public key is byte 0x52, is value for op code OP_2
5, the last byte is 0xa3, it is op code OP_CHECKMULTISIG

Now the stack for the script and the parsing stack is as following:
![bitcoin_script](https://github.com/wycl16514/golang-bitcoin-p2sh-contract/assets/7506958/0fcb4ec5-57e0-4f93-9674-a0caafc3d857)


Then the script execute the top command OP_2, this will result in pusing a value 2 on the stack:

![bitcoin_script (1)](https://github.com/wycl16514/golang-bitcoin-p2sh-contract/assets/7506958/99f9f8ac-7426-4180-b755-f3bf58eb6f44)


Because the first 3 elments on the script are two data elements, and one op code OP_2, executing then will have the following result:

![bitcoin_script (2)](https://github.com/wycl16514/golang-bitcoin-p2sh-contract/assets/7506958/762a0d9e-7537-499d-86c7-dee0da86a650)

This time when the script run the command OP_CHECKMULTISIG, it will take all elements on the parsing stack(m+n+3), if the execution result is success, it will push value 1 on the stack:

![bitcoin_script (4)](https://github.com/wycl16514/golang-bitcoin-p2sh-contract/assets/7506958/79c5a025-cced-4e48-8b0d-1e8709f9a87f)

This means the script verification is success. Let's see how to implement the evaluate process for p2sh transaction, it is a little be tricky, we need to check the scriptPubKey of to decide whether the current
transaction is kind of p2sh, if the scriptPubKey contains only three elements, the first one is OP_HASH160, the second is a chunk of data, the third one is OP_EQUAL, then the current transaction should be p2sh,
let's code this checking logic as following:
```g
func (b *BitcoinOpCode) opNum(op byte) bool {
	/*
		handle OP_0 to OP_16
	*/
	opNum := byte(0)
	if op >= OP_1 && op <= OP_16 {
		opNum = (op - OP_1) + 1
	}
	b.stack = append(b.stack, b.EncodeNum(int64(opNum)))
	return true
}

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
```
In aboved code, we get the ScriptPubKey of from the given output of previous transaction, and check its patter, if the script contains only three elements, the first element is OP_HASH160, the second one is 
a data chunk, the third one is OP_EQUAL, then the current transaction is p2sh kind. As we metionded above, if the transaction is p2sh, and when execute the OP_EQUAL operation and its result is 1, then we need
to parse the redeemscript and execute it is command, therefore we need to check this situation in the process of evaluation, and we handle it in the  BitcoinOpCode as following:
```g
const (
	/*
	   this is not a bitcoin script command, it is defined by ourself,
	   if we encounter the p2sh pattern on the script stack, that is the
	   fisrt element is data chunk, the second element is OP_HASH160,
	   the third element is a chunk of data, the fourth element is
	   OP_EQUAL, then we will use this command to do p2sh parsing
	*/
	OP_P2SH = 254
)

func (b *BitcoinOpCode) isP2sh() bool {
	/*
	   if we encounter the p2sh pattern on the script stack, that is the
	   fisrt element is data chunk, the second element is OP_HASH160,
	   the third element is a chunk of data, the fourth element is
	   OP_EQUAL
	*/
	if len(b.cmds[0]) != 1 || b.cmds[0][0] != OP_HASH160 {
		//the first element should be OP_HASH160
		return false
	}

	if len(b.cmds[1]) == 1 {
		//the second element should be hash data chunk
		return false
	}

	if len(b.cmds[0]) != 1 || b.cmds[0][0] != OP_EQUAL {
		//the third element should be OP_EQUAL
		return false
	}

	return true
}

func (b *BitcoinOpCode) opP2sh() bool {
	//the first command is OP_HASH160
	b.RemoveCmd()
	//the second element is a data chunk of hash
	h160 := b.RemoveCmd()
	//the third element is OP_EQUAL
	b.RemoveCmd()

	/*
	   the top element on stack is content of redeemscript, cache it and
	   do hash160 on it
	*/
	redeemScriptBinary := b.stack[len(b.stack)-1]
	if b.opHash160() != true {
		return false
	}
	//append the hash160 above onto the stack
	b.stack = append(b.stack, h160)
	//compare the two 160 hash on the stack
	if b.opEqual() != true {
		return false
	}

	if b.opVerify() != true {
		//if the two hash are equal, value 1 will push on the stack
		return false
	}

	//parse the redeemscript and append its command for handling
	scriptReader := bytes.NewReader(redeemScriptBinary)
	redeemScriptSig := NewScriptSig(bufio.NewReader(scriptReader))
	b.cmds = append(b.cmds, redeemScriptSig.cmds...)
	return true
}

func (b *BitcoinOpCode) AppendDataElement(element []byte) {
	b.stack = append(b.stack, element)
	/*
		we need to check the transaction is p2sh or not, if its p2sh,
		then the data element here is the binary data of reddemscript,
		and we insert OP_P2SH command to trigger the handling of the three
		p2sh command
	*/
	if b.isP2sh() {
		b.cmds = append([][]byte{[]byte{OP_P2SH}}, b.cmds...)
	}
}

func (b *BitcoinOpCode) ExecuteOperation(cmd int, z []byte) bool {
	/*
		if the operation executed successfuly return true, otherwise return false
	*/
	switch cmd {
        ...
        case OP_P2SH:
		return b.opP2sh()
        ...
        }
}

func (b *BitcoinOpCode) ExecuteOperation(cmd int, z []byte) bool {
	/*
		if the operation executed successfuly return true, otherwise return false
	*/
	switch cmd {
	case OP_CHECKSIG:
		return b.opCheckSig(z)
	case OP_DUP:
		return b.opDup()
	case OP_HASH160:
		return b.opHash160()
	case OP_EQUALVERIFY:
		return b.opEqualVerify()
	case OP_CHECKMULTISIG:
		return b.opCheckMultiSig(z)
	case OP_P2SH:
		return b.opP2sh()
	case OP_0:
		fallthrough
	case OP_1:
		fallthrough
	case OP_2:
		fallthrough
	case OP_3:
		fallthrough
	case OP_4:
		fallthrough
	case OP_5:
		fallthrough
	case OP_6:
		fallthrough
	case OP_7:
		fallthrough
	case OP_8:
		fallthrough
	case OP_9:
		fallthrough
	case OP_10:
		fallthrough
	case OP_11:
		fallthrough
	case OP_12:
		fallthrough
	case OP_13:
		fallthrough
	case OP_14:
		fallthrough
	case OP_15:
		fallthrough
	case OP_16:
		return b.opNum(byte(cmd))
	case OP_EQUAL:
		return b.opEqual()
	default:
		errStr := fmt.Sprintf("operation %s not implemented\n", b.opCodeNames[cmd])
		panic(errStr)
	}

	return false
}
```

In the above code, we defined a new op command OP_P2SH, this is not a official bitcoin script command, it is defined by ourself, every time we push a data element, we check the current op commands on the command
stack meet the pattern of p2sh scriptpubkey or not, if they are, then we insert the OP_P2SH command to the head of command stack, then in next time the command will be executed and the opP2sh method is called,
In this method, the top 3 elements on the parsing stack is OP_HASH160, data chunk of hash and OP_EQUAL, It removes the top 3 commands, then the top command on the top of parsing statck is the binary data of 
redeemscript, then it compute hash160 on the redeemscript, check its result is the same as the hash data chunk, if two hashes are the same, it parses the redeemscript and add commands of redeemscript for execution.



In the end of this section, let's see how to create wallet address for p2sh, when we create address for p2pkh transaction, we do hash160 on public key SEC format, then append prefix with 0x00 for mainnet or 0x6f 
for testnet, and then do base58 checksum encode, for p2sh address, the only difference is append 0x05 for mainnet and 0xc4 for testnet, the code is as following at point.go:
```g
func (p *Point) P2shAddress(compressed bool, testnet bool) string {
	hash160 := p.hash160(compressed)
	//if mainnet set first byte to 0x00 , 0x6f for testnet
	prefix := []byte{}
	if testnet {
		prefix = append(prefix, 0xc4)
	} else {
		prefix = append(prefix, 0x05)
	}
	//do base58 checksum
	return Base58Checksum(append(prefix, hash160...))
}

```



