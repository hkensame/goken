package persister

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"

	"github.com/hkensame/goken/pkg/log"
)

type blockReader struct {
	bm  *BlockManager
	now int
	end int
}

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

// readBlock会自动seek到指定位置
func (bm *BlockManager) readBlock(blocknum int) ([]byte, error) {
	if err := bm.seekBlock(blocknum); err != nil {
		return nil, err
	}
	reader := bufio.NewReaderSize(bm.File, BlockSize)
	return bm.read(reader, BlockSize)
}

func (bm *BlockManager) seekBlock(blocknum int) error {
	off := blocknum * BlockSize
	if err := bm.seek(off, io.SeekStart); err != nil {
		return err
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
	return bm.readBlock(int(bm.md.UsagedBlockNum))
}

func (bm *BlockManager) SeekBlock() error {
	return bm.seekBlock(int(bm.md.UsagedBlockNum))
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

// 这个函数会返回一个blockReader用于读取每个block内的所有entries
func (bm *BlockManager) ReadBlockEntries() *blockReader {
	return &blockReader{
		bm:  bm,
		now: 1,
		end: int(bm.md.UsagedBlockNum),
	}
}

// 这个函数每次调用时都会先偏移到对应的block中,故可以在blockReader读取之前的块时在最新块中添加entry,
// 但不要在br读取最新块的entries时写新的entry
func (br *blockReader) Next() ([][]byte, error) {
	if br.now > br.end {
		return nil, nil
	}
	b, err := br.bm.readBlock(br.now)
	if err != nil {
		return nil, err
	}
	block := &Block{}
	if err := block.Unmarshal(b); err != nil {
		return nil, err
	}
	res := block.GetEntries()
	br.now++
	return res, nil
}

// 判断现在使用的Block是否还够用,不够用就换新的Block
// 这样可能会造成每个block都有可能会有几个到几百个字节未使用,但保证了每个条目都落在一个块下
func (bm *BlockManager) CheckStatus(needSize int) error {
	//如果容量已经不足则先落盘数据
	if bm.UsagedBlock.RemainedSize()-needSize < 0 {
		if err := bm.WriteBlock(); err != nil {
			return err
		}
		bm.md.UsagedBlockNum++
		//需要扩充文件
		if bm.md.UsagedBlockNum > bm.md.BlockNums {
			bm.md.BlockNums *= 2
			if err := bm.Expansion(int(bm.md.BlockNums)); err != nil {
				return err
			}
		}
		if err := bm.StoreHeaderData(); err != nil {
			return err
		}
		if err := bm.SeekBlock(); err != nil {
			return err
		}
		//清空数组
		bm.UsagedBlock.EntriesData = [BodySize]byte{}
		bm.UsagedBlock.Header.EntryNums = 0
		bm.UsagedBlock.Header.BlockNum = int16(bm.md.UsagedBlockNum)
		bm.UsagedBlock.Header.UsedSize = 0
		//bm.dirty = true
	}
	return nil
}

func (bm *BlockManager) WriteEntry(d []byte) error {
	be := NewBlockEntry(d)
	if err := bm.CheckStatus(int(be.EntrySize + 2)); err != nil {
		return err
	}
	bm.dirty = true
	bm.UsagedBlock.Header.EntryNums++
	copy(bm.UsagedBlock.EntriesData[bm.UsagedBlock.Header.UsedSize:], be.Encode())
	bm.UsagedBlock.Header.UsedSize += be.EntrySize + 2
	return nil
}

// 要求必须写成功到磁盘中
func (bm *BlockManager) MustWriteEntry(d []byte) error {
	if err := bm.WriteBlock(); err != nil {
		return err
	}
	return bm.Flush()
}
