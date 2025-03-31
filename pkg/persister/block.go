package persister

import (
	"encoding/binary"
	"os"
	"unsafe"

	"github.com/hkensame/goken/pkg/errors"
)

const (
	BlockSize int = 1 << 12
	PageSize  int = 1 << 14

	//注意headerSize包括了8个字节的checksum
	HeaderSize = int(unsafe.Sizeof(BlockHeader{}) + 8)
	BodySize   = BlockSize - HeaderSize
)

var (
	ErrPersistFailed   = errors.New("数据持久到磁盘失败")
	ErrAllocateFailed  = errors.New("文件扩容失败")
	ErrSeekFailed      = errors.New("文件寻址失败")
	ErrWriteFailed     = errors.New("数据写入文件失败")
	ErrReadFailed      = errors.New("数据读取失败")
	ErrUnmarshalFailed = errors.New("给定的二进制数据反序列化失败")
	ErrChecksumFailed  = errors.New("读出的数据可能有误,无法经过checksum")
)

type BlockEntry struct {
	EntrySize int16
	Data      []byte
}

type BlockHeader struct {
	// 已经使用的字节数
	UsedSize int16
	// 对应的块号
	BlockNum int16
	// // block的元数据大小
	// BlockHeaderSize int16
	// // block的body数据大小
	// BlockBodySize int16
	// 现在已经存储的条目数
	EntryNums int16
	Align     int16
}

// 因为block总共被固定为4k字节,所以里面的字段使用int16是安全的
type Block struct {
	Header BlockHeader
	// 两个checksum
	HeaderCheckSum uint32
	BodyCheckSum   uint32
	EntriesData    [BodySize]byte
}

type BlockManager struct {
	File *os.File
	//起始给一个文件分配多少个块,但注意如果分配的所有块都被使用了还是会在新文件中分配新块
	BlockNums int32
	//读取时会把老的block读入到这个切片中
	Blocks []Block

	UsagedBlock *Block
	//正在使用的块号,下标从1开始计起
	UsagedBlockNum int32
}

func NewBlockEntry(b []byte) *BlockEntry {
	return &BlockEntry{
		Data:      b,
		EntrySize: int16(len(b)),
	}
}

func (be *BlockEntry) Encode() []byte {
	be.EntrySize = int16(len(be.Data))
	buf := make([]byte, 2+be.EntrySize)
	binary.LittleEndian.PutUint16(buf[:2], uint16(be.EntrySize))
	copy(buf[2:], be.Data)
	return buf
}

// NOTICE 未来应该添加设置,允许制定不同一致性的写管理器,这里先写一个带buff的试试水
// 注意这个函数没有先读取之前的文件,我这里留了一个点就是第一个block用于存储元数据
func MustNewBlockManager(path string, blocks int) *BlockManager {
	bm := &BlockManager{
		BlockNums:   int32(blocks),
		UsagedBlock: &Block{},
	}

	var err error
	bm.File, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	if err := bm.Expansion(blocks); err != nil {
		panic(err)
	}

	bm.UsagedBlockNum = 1
	bm.UsagedBlock.Header = BlockHeader{
		UsedSize: 0,
		BlockNum: 1,
		// BlockHeaderSize: int16(HeaderSize) + 8,
		// BlockBodySize:   int16(BlockSize) - int16(HeaderSize) - 8,
	}

	if err := bm.WriteBlock(); err != nil {
		panic(err)
	}

	return bm
}

// 判断现在使用的Block是否还够用,不够用就换新的Block
// 这样可能会造成每个block都有可能会有几个到几百个字节未使用,但保证了每个条目都落在一个块下
func (bm *BlockManager) CheckStatus(needSize int) error {
	//如果容量已经不足则先落盘数据
	if bm.UsagedBlock.RemainedSize()-needSize < 0 {
		if err := bm.WriteBlock(); err != nil {
			return err
		}
		bm.UsagedBlockNum++
		//需要扩充文件
		if bm.UsagedBlockNum > bm.BlockNums {
			bm.BlockNums *= 2
			if err := bm.Expansion(int(bm.BlockNums)); err != nil {
				return err
			}
		}

		if err := bm.SeekBlock(); err != nil {
			return err
		}
		//清空数组
		bm.UsagedBlock.EntriesData = [BodySize]byte{}
		bm.UsagedBlock.Header.EntryNums = 0
		bm.UsagedBlock.Header.BlockNum = int16(bm.UsagedBlockNum)
		bm.UsagedBlock.Header.UsedSize = 0
	}
	return nil
}

func (bm *BlockManager) WriteEntry(d []byte) error {
	be := NewBlockEntry(d)
	if err := bm.CheckStatus(int(be.EntrySize)); err != nil {
		return err
	}
	bm.UsagedBlock.Header.EntryNums++
	copy(bm.UsagedBlock.EntriesData[bm.UsagedBlock.Header.UsedSize:], be.Encode())
	bm.UsagedBlock.Header.UsedSize += be.EntrySize
	return nil
}
