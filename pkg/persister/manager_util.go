package persister

import (
	"bufio"
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
	// if err := bm.File.Sync(); err != nil {
	// 	panic(err)
	// }
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

// 这两个函数只负责UsedBlock
func (bm *BlockManager) WriteBlock() error {
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

func (bm *BlockManager) seek(offset int, whence int) error {
	_, err := bm.File.Seek(int64(offset), whence)
	if err != nil {
		return ErrSeekFailed
	}
	return nil
}

func (bm *BlockManager) SeekBlock() error {
	off := int(bm.UsagedBlockNum) * BlockSize
	if err := bm.seek(off, io.SeekStart); err != nil {
		return err
	}
	return nil
}

func (bm *BlockManager) Expansion(blocknums int) error {
	if blocknums <= int(bm.BlockNums) {
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
