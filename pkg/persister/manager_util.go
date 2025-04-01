package persister

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"

	"github.com/hkensame/goken/pkg/log"
)

func (bm *BlockManager) write(writer io.Writer, b []byte) error {
	var written int
	for written < len(b) {
		n, err := writer.Write(b[written:])
		if err != nil {
			log.Errorf("[persister] 数据写入失败 err = %v", err)
			return ErrWriteFailed
		}
		written += n
	}
	return nil
}

func (bm *BlockManager) read(reader io.Reader, size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := io.ReadFull(reader, b)
	if err != nil {
		log.Errorf("[persister] 数据读取失败 err = %v", err)
		return nil, ErrReadFailed
	}
	return b, nil
}

func (bm *BlockManager) seek(offset int, whence int) error {
	_, err := bm.File.Seek(int64(offset), whence)
	if err != nil {
		log.Errorf("[persister] 数据寻址失败 err = %v", err)
		return ErrSeekFailed
	}
	return nil
}

func (bm *BlockManager) sync() error {
	if err := bm.File.Sync(); err != nil {
		log.Errorf("[persister] 数据同步失败 err = %v", err)
		return ErrSyncFailed
	}
	return nil
}

// 这两个函数只负责UsedBlock
func (bm *BlockManager) WriteBlock() error {
	if !bm.dirty {
		return nil
	}
	if err := bm.SeekBlock(); err != nil {
		return err
	}
	writer := bufio.NewWriterSize(bm.File, BlockSize)
	defer writer.Flush()
	bm.UsagedBlock.SetCheckSum()
	return bm.write(writer, bm.UsagedBlock.Marshal())
}

func (bm *BlockManager) ReadBlock() ([]byte, error) {
	if err := bm.SeekBlock(); err != nil {
		return nil, err
	}
	reader := bufio.NewReaderSize(bm.File, BlockSize)
	return bm.read(reader, BlockSize)
}

func (bm *BlockManager) SeekBlock() error {
	off := int(bm.md.UsagedBlockNum) * BlockSize
	if err := bm.seek(off, io.SeekStart); err != nil {
		return err
	}
	return nil
}

func (bm *BlockManager) Expansion(blocknums int) error {
	if blocknums <= int(bm.md.BlockNums) {
		return nil
	}
	if err := bm.File.Truncate(int64(blocknums+1) * int64(BlockSize)); err != nil {
		return ErrAllocateFailed
	}
	return nil
}

func (bm *BlockManager) Flush() error {
	if err := bm.WriteBlock(); err != nil {
		return err
	}
	return nil
}

func (bm *BlockManager) StoreHeaderData() error {
	buf := make([]byte, 0)
	//buf = binary.LittleEndian.AppendUint32(buf, uint32(HeaderBlockSize))
	buf = binary.LittleEndian.AppendUint32(buf, uint32(bm.md.BlockNums))
	buf = binary.LittleEndian.AppendUint32(buf, uint32(bm.md.UsagedBlockNum))

	if err := bm.seek(0, io.SeekStart); err != nil {
		return err
	}
	if err := bm.write(bm.File, buf); err != nil {
		return err
	}
	return bm.sync()
}

func (bm *BlockManager) LoadHeaderData() error {
	buf := make([]byte, HeaderBlockSize)

	if err := bm.seek(0, io.SeekStart); err != nil {
		return err
	}
	if _, err := io.ReadFull(bm.File, buf); err != nil {
		return err
	}

	b := bytes.NewBuffer(buf)

	binary.Read(b, binary.LittleEndian, bm.md)
	data, err := bm.ReadBlock()
	if err != nil {
		return err
	}
	return bm.UsagedBlock.Unmarshal(data)
}
