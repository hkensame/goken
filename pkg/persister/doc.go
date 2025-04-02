package persister

/*
	该包暂时没有开放接口允许使用者重新写老block

	后续想提高功能性的话可以多以下几个字段:
	1. bm.nextUsefulBlockNum int32 下一个可使用的blockNum,通过这个字段来分配块号
	2. b.nextBlockNum int32 下一个同类型的块号,通过这个字段可以实现多块存储数据
	3. b.firstDeleteEntryAt int16 第一条删除的entry的下标位置,且每个entry加入时而外用一个bit(从size里拿)来判断是否删除
		因为一个block最多有4096字节,所以size虽有16bit但是用不完所有bit

*/
