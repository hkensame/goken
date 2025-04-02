package persister

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/hkensame/goken/pkg/errors"
)

//Header Block的存储默认是立马刷新而不使用缓存

func (bm *BlockManager) storeHeaderBlock(other []byte) error {
	buf := make([]byte, 0, HeaderBlockSize+len(other))
	buf = binary.LittleEndian.AppendUint32(buf, uint32(bm.md.BlockNums))
	buf = binary.LittleEndian.AppendUint32(buf, uint32(bm.md.UsagedBlockNum))
	//如果other不为空就加入到头block中
	if other != nil && len(other) != 0 {
		if len(other) > BlockSize-HeaderBlockSize-2 {
			return errors.New("需要存储的额外信息太多,persister暂时不支持")
		}
		buf = binary.LittleEndian.AppendUint16(buf, uint16(len(other)))
		buf = append(buf, other...)
	}

	if err := bm.seek(0, io.SeekStart); err != nil {
		return err
	}
	if err := bm.write(bm.File, buf); err != nil {
		return err
	}
	return bm.sync()
}

func (bm *BlockManager) LoadHeaderData() error {
	if err := bm.seek(0, io.SeekStart); err != nil {
		return err
	}
	buf, err := bm.read(bm.File, HeaderBlockSize)
	if err != nil {
		return err
	}

	b := bytes.NewBuffer(buf)

	binary.Read(b, binary.LittleEndian, bm.md)
	data, err := bm.ReadBlock()
	if err != nil {
		return err
	}

	bm.UsagedBlock.Unmarshal(data)
	return nil
}

func (bm *BlockManager) GetCustomData() ([]byte, error) {
	if err := bm.seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	buf, err := bm.read(bm.File, HeaderBlockSize+2)
	if err != nil {
		return nil, err
	}

	var otherSize uint16
	var other []byte = nil
	otherSize = binary.LittleEndian.Uint16(buf[HeaderBlockSize:])

	if otherSize != 0 {
		other, err = bm.read(bm.File, int(otherSize))
		if err != nil {
			return nil, err
		}
	}
	return other, nil
}

// other可以是额外传入的自定义配置信息
func (bm *BlockManager) StoreHeaderData() error {
	return bm.storeHeaderBlock(nil)
}

func (bm *BlockManager) StoreCustomData(other []byte) error {
	return bm.storeHeaderBlock(other)
}
