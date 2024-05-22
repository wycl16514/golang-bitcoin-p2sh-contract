package transaction

import (
	"bufio"
	"io"
	"math/big"
)

type ScriptSig struct {
	cmds          [][]byte
	bitcoinOpCode *BitcoinOpCode
}

const (
	//[0x1, 0x4b] -> [1, 75]
	SCRIPT_DATA_LENGTH_BEGIN = 1
	SCRIPT_DATA_LENGTH_END   = 75
	OP_PUSHDATA1             = 76
	OP_PUSHDATA2             = 77
)

func InitScriptSig(cmds [][]byte) *ScriptSig {
	bitcoinOpCode := NewBicoinOpCode()
	bitcoinOpCode.cmds = cmds
	return &ScriptSig{
		bitcoinOpCode: bitcoinOpCode,
	}
}

func NewScriptSig(reader *bufio.Reader) *ScriptSig {
	cmds := [][]byte{}
	/*
		At the beginning is the total length for script field
	*/
	scriptLen := ReadVarint(reader).Int64()
	count := int64(0)
	current := make([]byte, 1)
	var current_byte byte
	for count < scriptLen {
		//buf fix
		//reader.Read(current)
		io.ReadFull(reader, current)
		//operation
		count += 1
		current_byte = current[0]
		if current_byte >= SCRIPT_DATA_LENGTH_BEGIN &&
			current_byte <= SCRIPT_DATA_LENGTH_END {
			//push the following bytes of data onto stack
			data := make([]byte, current_byte)
			//bug fix
			//reader.Read(data)
			io.ReadFull(reader, data)
			cmds = append(cmds, data)
			count += int64(current_byte)
		} else if current_byte == OP_PUSHDATA1 {
			/*
				read the following byte as the length of data
			*/
			length := make([]byte, 1)
			//bug fix
			//reader.Read(length)
			io.ReadFull(reader, length)

			data := make([]byte, length[0])
			//reader.Read(data)
			io.ReadFull(reader, data)
			cmds = append(cmds, data)
			count += int64(length[0] + 1)
		} else if current_byte == OP_PUSHDATA2 {
			/*
				read the following 2 bytes as length of data
			*/
			lenBuf := make([]byte, 2)
			//buf ifx
			//reader.Read(lenBuf)
			io.ReadFull(reader, lenBuf)
			length := LittleEndianToBigInt(lenBuf, LITTLE_ENDIAN_2_BYTES)
			data := make([]byte, length.Int64())
			//reader.Read(data)
			io.ReadFull(reader, data)
			cmds = append(cmds, data)
			count += int64(2 + length.Int64())
		} else {
			//is data processing instruction
			cmds = append(cmds, []byte{current_byte})
		}
	}

	if count != scriptLen {
		panic("parsing script field fail")
	}

	return InitScriptSig(cmds)
}

func (s *ScriptSig) Evaluate(z []byte) bool {
	for s.bitcoinOpCode.HasCmd() {
		cmd := s.bitcoinOpCode.RemoveCmd()
		if len(cmd) == 1 {
			//this is op code, run it
			opRes := s.bitcoinOpCode.ExecuteOperaion(int(cmd[0]), z)
			if opRes != true {
				return false
			}
		} else {
			s.bitcoinOpCode.AppendDataElement(cmd)
		}
	}

	/*
		After running all the operations in the scripts and the stack is empty,
		then evaluation fail, otherwise we check the top element of the stack,
		if it value is 0, then fail, if the value is not 0, then success
	*/
	if len(s.bitcoinOpCode.stack) == 0 {
		return false
	}
	if len(s.bitcoinOpCode.stack[len(s.bitcoinOpCode.stack)-1]) == 0 {
		return false
	}

	return true
}

func (s *ScriptSig) rawSerialize() []byte {
	result := []byte{}
	for _, cmd := range s.bitcoinOpCode.cmds {
		if len(cmd) == 1 {
			//only one byte means its an instruction
			result = append(result, cmd...)
		} else {
			length := len(cmd)
			if length <= SCRIPT_DATA_LENGTH_END {
				//length in [0x01, 0x4b]
				result = append(result, byte(length))
			} else if length > SCRIPT_DATA_LENGTH_END && length < 0x100 {
				//this is OP_PUSHDATA1 command,
				//push the command and then the next byte is the length of the data
				result = append(result, OP_PUSHDATA1)
				result = append(result, byte(length))
			} else if length >= 0x100 && length <= 520 {
				/*
					this is OP_PUSHDATA2 command, we push the command
					and then two byte for the data length but in little endian format
				*/
				result = append(result, OP_PUSHDATA2)
				lenBuf := BigIntToLittleEndian(big.NewInt(int64(length)), LITTLE_ENDIAN_2_BYTES)
				result = append(result, lenBuf...)
			} else {
				panic("too long an cmd")
			}

			//append the chunk of data with given length
			result = append(result, cmd...)
		}
	}

	return result
}

func (s *ScriptSig) Serialize() []byte {
	rawResult := s.rawSerialize()
	total := len(rawResult)
	result := []byte{}
	//encode the total length of script at the head
	result = append(result, EncodeVarint(big.NewInt(int64(total)))...)
	result = append(result, rawResult...)
	return result
}

func (s *ScriptSig) Add(script *ScriptSig) *ScriptSig {
	cmds := make([][]byte, 0)
	cmds = append(cmds, s.bitcoinOpCode.cmds...)
	cmds = append(cmds, script.bitcoinOpCode.cmds...)
	return InitScriptSig(cmds)
}
