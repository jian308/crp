/*
 * @Author: 杨小灿jian308@qq.com
 * @Date: 2023-04-05 17:27:21
 * @LastEditors: 杨小灿jian308@qq.com
 * @LastEditTime: 2023-04-09 23:09:46
 */
package unpack

import (
	"encoding/binary"
	"errors"
	"io"
)

var MsgToken = ""

// Encode
func Encode(bytesBuffer io.Writer, content []byte) error {
	if err := binary.Write(bytesBuffer, binary.BigEndian, []byte(MsgToken)); err != nil {
		return err
	}
	clen := int32(len(content))
	if err := binary.Write(bytesBuffer, binary.BigEndian, clen); err != nil {
		return err
	}
	if err := binary.Write(bytesBuffer, binary.BigEndian, content); err != nil {
		return err
	}
	return nil
}

// Decode
func Decode(bytesBuffer io.Reader) (bodyBuf []byte, err error) {
	headBuf := make([]byte, len(MsgToken))
	//log.Debug(cap(headBuf), MsgToken)
	if _, err := io.ReadFull(bytesBuffer, headBuf); err != nil {
		return nil, err
	}
	if string(headBuf) != MsgToken {
		return nil, errors.New("token不对")
	}
	lbuf := make([]byte, 4)
	if _, err := io.ReadFull(bytesBuffer, lbuf); err != nil {
		return nil, err
	}

	l := binary.BigEndian.Uint32(lbuf)
	body := make([]byte, l)
	if _, err := io.ReadFull(bytesBuffer, body); err != nil {
		return nil, err
	}
	return body, err
}
