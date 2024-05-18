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


