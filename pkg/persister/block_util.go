package persister

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
)

func (b *Block) RemainedSize() int {
	return BodySize - int(b.Header.UsedSize)
}

// 返回的下标是还未使用的body的第一个字节的seek地址
func (b *Block) At() int {
	offset := int(b.Header.BlockNum * int16(BlockSize))
	return int(b.Header.UsedSize+int16(HeaderSize)) + offset
}

func (b *Block) Marshal() []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	binary.Write(buf, binary.LittleEndian, b)
	return buf.Bytes()
}

// 可以进行checksum检查
func (b *Block) Unmarshal(data []byte) error {
	//如果连一个头都读不到则退出
	if len(data) < HeaderSize {
		return ErrUnmarshalFailed
	}
	buf := bytes.NewReader(data[:HeaderSize])

	if err := binary.Read(buf, binary.LittleEndian, &b.Header); err != nil {
		return ErrUnmarshalFailed
	}
	//如果data里的数据没有满足合适的大小就退出
	if len(data) < HeaderSize+int(b.Header.UsedSize) {
		return ErrUnmarshalFailed
	}
	binary.Read(buf, binary.LittleEndian, &b.HeaderCheckSum)
	binary.Read(buf, binary.LittleEndian, &b.BodyCheckSum)

	//一定成功
	copy(b.EntriesData[:], data[HeaderSize:])

	if !b.CheckSum() {
		return ErrChecksumFailed
	}
	return nil
}

func (b *Block) SetCheckSum() {
	headbuf := bytes.NewBuffer(make([]byte, 0))
	bodybuf := bytes.NewBuffer(make([]byte, 0))

	binary.Write(headbuf, binary.LittleEndian, b.Header)
	b.HeaderCheckSum = crc32.ChecksumIEEE(headbuf.Bytes())
	binary.Write(bodybuf, binary.LittleEndian, b.EntriesData)
	b.BodyCheckSum = crc32.ChecksumIEEE(bodybuf.Bytes())
}

func (b *Block) CheckSum() bool {
	headbuf := bytes.NewBuffer(make([]byte, 0))
	bodybuf := bytes.NewBuffer(make([]byte, 0))

	binary.Write(headbuf, binary.LittleEndian, b.Header)
	if b.HeaderCheckSum != crc32.ChecksumIEEE(headbuf.Bytes()) {
		return false
	}
	binary.Write(bodybuf, binary.LittleEndian, b.EntriesData)
	if b.BodyCheckSum != crc32.ChecksumIEEE(bodybuf.Bytes()) {
		return false
	}

	return true
}
