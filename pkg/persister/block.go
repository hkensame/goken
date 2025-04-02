package persister

import (
	"encoding/binary"
	"os"
	"unsafe"

	"github.com/hkensame/goken/pkg/errors"
	"github.com/hkensame/goken/pkg/log"
)

const (
	BlockSize int = 1 << 12
	PageSize  int = 1 << 14

	//注意headerSize包括了8个字节的checksum
	HeaderSize      = int(unsafe.Sizeof(BlockHeader{}) + 8)
	BodySize        = BlockSize - HeaderSize
	HeaderBlockSize = 8
)

var (
	ErrPersistFailed   = errors.New("数据持久到磁盘失败")
	ErrAllocateFailed  = errors.New("文件扩容失败")
	ErrSeekFailed      = errors.New("文件寻址失败")
	ErrSyncFailed      = errors.New("数据同步到文件失败")
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

type metadata struct {
	//起始给一个文件分配多少个块,但注意如果分配的所有块都被使用了还是会在新文件中分配新块
	BlockNums int32
	//正在使用的块号,下标从1开始计起
	UsagedBlockNum int32
}

type BlockManager struct {
	File *os.File

	md *metadata

	//读取时会把老的block读入到这个切片中
	Blocks      []Block
	UsagedBlock *Block

	dirty bool
}

func NewBlockEntry(b []byte) *BlockEntry {
	return &BlockEntry{
		Data:      b,
		EntrySize: int16(len(b)),
	}
}

func (be *BlockEntry) Encode() []byte {
	buf := make([]byte, 2+be.EntrySize)
	binary.LittleEndian.PutUint16(buf[:2], uint16(be.EntrySize))
	copy(buf[2:], be.Data)
	return buf
}

// 写了一个脆弱的读取系统,不要修改文件内的内容
// 必知:该persister提供了基本的写block和读block,如果希望强一致性就在每次写的时候调用flush
// 整个包不提供锁和并发安全保障,请将persister当做一种需要锁的资源
func MustNewBlockManager(path string, blocks int) *BlockManager {
	bm := &BlockManager{
		md:          new(metadata),
		UsagedBlock: &Block{},
	}
	bm.md.BlockNums = int32(blocks)

	var err error
	bm.File, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	//如果文件是第一次写入或者数据不对就刷新
	if err := bm.LoadHeaderData(); err != nil {
		log.Warnf("读取文件header block失败 warn = %v", err)
		if err := bm.Expansion(blocks); err != nil {
			panic(err)
		}

		bm.md.UsagedBlockNum = 1
		bm.UsagedBlock.Header = BlockHeader{
			UsedSize: 0,
			BlockNum: 1,
		}
		bm.dirty = true
		if err := bm.WriteBlock(); err != nil {
			panic(err)
		}
		if err := bm.StoreHeaderData(); err != nil {
			panic(err)
		}
	}
	return bm
}
