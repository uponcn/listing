package listing

import (
	"os"
	"fmt"
	"sync"
	"time"
	"bytes"
	"math/rand"
	"encoding/binary"
)

var SEP = string(os.PathSeparator)
var ROOT = "." + SEP + "data" + SEP
var STACK *sync.RWMutex = new(sync.RWMutex)

func Show(){
	fp, _ := os.OpenFile(ROOT + "stack.s", os.O_RDONLY, 0777)
	defer fp.Close()
	fp.Seek( 256, os.SEEK_SET )
	buf := make([]byte, 256)
	pos := 256
	nums := make( []int64, 32 )
	for{
		n, _ := fp.Read(buf)
		if n!=256 {
			break
		}
		Scan( buf, &nums )
		fmt.Println( pos, ":", nums )
		pos += 256
	}
}

// reset data files
func StackClean() {
	os.WriteFile(ROOT+"stack.s", bytes.Repeat([]byte{0,0,0,0,0,0,0,0}, 32), 0777)
	os.WriteFile(ROOT+"stackFree.s", []byte{0,0,0,8,0,0,0,8}, 0777)
}

//stackGetNum 计算size字节需要多少blockByte块，num: 总块数，level: 层数（包含数据层）
func stackGetNum(size int) (num, level int) {
	n := size >> 8
	if size&0xFF > 0 {
		n++
	}
	num = n
	level = 0
	for n > 1 {
		if n&0x1F > 0 {
			n = 1 + (n >> 5)
		} else {
			n = n >> 5
		}
		num += n
		level++
	}
	return
}

//stackLevel 获得指定位置的层级结构
func stackLevel(num int) []int {
	if num == 0 {
		return []int{0}
	}
	res := []int{ num&0xFF }
	num >>= 8
	for num>0 {
		res = append(res, num&0x1F)
		num >>= 5
	}
	return res
}

//stackPos 获得指定位置的位置结构
func stackPos(pos, num int) []int {
	poss := make([]int, 0)
	fp, err := os.OpenFile(ROOT + "stack.s", os.O_RDONLY, 0777)
	defer fp.Close()
	if err==nil {
		levs := stackLevel(num)
		lens := len(levs)
		buf := make([]byte, 8)
		for lens > 0 {
			lens--
			if lens > 0 {
				pos += levs[lens] << 3
				poss = append(poss, pos)
				fp.Seek(int64(pos), os.SEEK_SET)
				fp.Read(buf)
				pos = int(Uint64Big(buf))
			} else {
				pos += levs[0]
				poss = append(poss, pos)
			}
		}
	}
	return poss
}

//stackPosRand 打乱次序
func stackPosRand(poss []int64) []int64 {
	lens := len(poss)
	rand.Seed(time.Now().UnixNano())
	arr := rand.Perm(lens)
	res := make([]int64, lens)
	for i := 0; i < lens; i++ {
		res[i] = poss[arr[i]]
	}
	return res
}

//get num blocks 
func stackGetBlocks( num uint32 ) []int64 {
	blocks := make( []int64, num )
	STACK.Lock()
	fp, err := os.OpenFile(ROOT + "stackFree.s", os.O_RDWR, 0777)
	if err==nil {
		buf := make( []byte, 8 )
		fp.Read( buf )
		start := Uint32Big( buf[0:4] )
		end   := Uint32Big( buf[4:8] )
		n := (end-start)>>3
		if n>=num {
			fp.Seek( int64(start), os.SEEK_SET )
			for i,_ := range blocks {
				fp.Read( buf )
				blocks[i] = int64( Uint64Big(buf) )
			}
			fp.Seek( 0, os.SEEK_SET )
			if n==num {
				fp.Write( []byte{ 0,0,0,8,0,0,0,8 } )
				fp.Truncate( 8 )
			}else{
				fp.Write( VarsToBytes( start+(num<<3) ) )
			}
		}else{
			if n>0 {
				bufn := make([]byte, n<<8)
				fp.Seek( int64(start), os.SEEK_SET )
				fp.Read(bufn)
				binary.Read( bytes.NewBuffer(bufn), binary.BigEndian, blocks[:n] )
				fp.Seek( 0, os.SEEK_SET )
				fp.Write( []byte{ 0,0,0,8,0,0,0,8 } )
				fp.Truncate( 8 )
			}
			fp2, err := os.OpenFile(ROOT + "stack.s", os.O_WRONLY, 0777)
			if err==nil {
				pos, err := fp2.Seek( 0, os.SEEK_END )
				if err==nil {
					if (pos&0xFF)>0 {
						fp2.Write( make([]byte, 256-(pos&0xFF)) )
					}
					fp2.Write( make([]byte, (num-n)<<8 ) )
				}
				fp2.Close()
				for i:=n; i<num; i++ {
					blocks[i] = pos
					pos += 256
				}
			}
		}
	}
	fp.Close()
	STACK.Unlock()
	return blocks
}

func stackRangeBlock(from, to, pos, space int64) (blocks []int64, start, end int64) {
	var nums, bits, pf, pt int64 = 0x7C00000000000000, 58, 0, 0
	for (nums&int64(space-1))==0 {
		nums >>= 5
		bits -= 5
	}
	STACK.RLock()
	fp, err := os.OpenFile(ROOT + "stack.s", os.O_RDONLY, 0777)
	lens := 0
	if err==nil {
		blocks = []int64{ pos }
		for nums>0xFF {
			poss := make( []int64, 0 )
			lens = len(blocks)
			pf = from&nums
			pt = ((to-1)&nums)+ (1<<bits)
			fp.Seek( blocks[0]+(pf>>(bits-3)), os.SEEK_SET )
			if lens==1 {
				bufn := make( []byte, (pt-pf)>>(bits-3) )
				fp.Read(bufn)
				rangeX := make( []int64, (pt-pf)>>bits )
				binary.Read( bytes.NewBuffer(bufn), binary.BigEndian, &rangeX )
				poss = append(poss, rangeX...)
			}else{
				bufn := make( []byte, 256-(pf>>(bits-3)) )
				fp.Read(bufn)
				rangeX := make( []int64, 32-(pf>>bits) )
				binary.Read( bytes.NewBuffer(bufn), binary.BigEndian, &rangeX )
				poss = append(poss, rangeX...)
				bufn = make( []byte, 256 )
				for i:=1; i<lens-1; i++ {
					rangeY := make([]int64, 32)
					fp.Seek( blocks[i], os.SEEK_SET )
					fp.Read(bufn)
					binary.Read( bytes.NewBuffer(bufn), binary.BigEndian, &rangeY )
					poss = append( poss, rangeY... )
				}
				fp.Seek( blocks[lens-1], os.SEEK_SET )
				bufn = make( []byte, pt>>(bits-3) )
				fp.Read(bufn)
				rangeZ := make( []int64, pt>>8 )
				binary.Read( bytes.NewBuffer(bufn), binary.BigEndian, &rangeZ )
				poss = append( poss, rangeZ... )
			}
			blocks = poss
			nums >>= 5
			bits -= 5
		}
	}
	fp.Close()
	STACK.RUnlock()
	start, end = from&0xFF, to&0xFF
	return 
}

func stackRangeWrite( blocks []int64, data []byte, start, end int64 ) int{
	lens := len(blocks)
	STACK.Lock()
	fp, _ := os.OpenFile(ROOT + "stack.s", os.O_WRONLY, 0777)
	fp.Seek( blocks[0]+start, os.SEEK_SET )
	if lens==1 {
		fp.Write( data )
	}else{
		fp.Write( data[0:256-start] )
		n := len(blocks)-1
		index := 256-start
		for i:=1;i<n;i++ {
			fp.Seek( blocks[i], os.SEEK_SET )
			fp.Write( data[index:index+256] )
			index += 256
		}
		fp.Seek( blocks[n], os.SEEK_SET )
		fp.Write( data[index:] )
	}
	fp.Close()
	STACK.Unlock()
	return len(data)
}
func stackEnlarge(pos, size, space, spaceNew int) []byte {
	if (spaceNew&0xFF) != 0 {
		spaceNew += 256 - (spaceNew&0xFF)
	}
	poss := stackPos(pos, space-1)
	num2, level2 := stackGetNum(spaceNew)
	num1, level1 := stackGetNum(space)
	blocks := stackGetBlocks( uint32(num2-num1) )
	num := (spaceNew-space)>>8
	data := make([]byte, 0)
	if space==256 {
		data = VarsToBytes(pos)
	}
	for i:=0; i<num; i++ {
		data = append(data, VarsToBytes( blocks[i] )... )
	}
	var n1 int
	STACK.Lock()
	fp, _ := os.OpenFile(ROOT + "stack.s", os.O_WRONLY, 0777)
	for n2:=level2-1;n2>=0;n2-- {
		n1 = level1 + n2 -level2
		if n1 >= 0 {
			posS := int64(poss[n1]&0xFF)+8
			if posS != 0 {
				dataLen := len(data)
				fp.Seek( int64(poss[n1]+8),  os.SEEK_SET )
				if 256-int(posS)>dataLen {
					fp.Write(data)
					data = []byte{}
				}else{
					fp.Write( data[0:256-posS] )
					data = data[256-posS:]
				}
			}
		}
		dataLen := len(data)
		dat := make([]byte, 0)
		if n1==0 {
			dat = VarsToBytes(pos)
		}
		for i:=0; i<dataLen; i+=256 {
			
			dat = append(dat, VarsToBytes( blocks[num] )...)
			fp.Seek( blocks[num],  os.SEEK_SET )
			num++
			if i+256>dataLen {
				fp.Write( data[i:dataLen] )
			}else{
				fp.Write( data[i:i+256] )
			}
		}
		if len(dat)==0 {
			break
		}
		data = dat
	}
	fp.Close()
	STACK.Unlock()
	posNew := pos
	if level2>level1 {
		posNew = int(blocks[len(blocks)-1])
	}
	return VarsToBytes(uint16(0xFF00), posNew, size, spaceNew )
}

//StackNew 
func StackNew(data []byte) []byte {
	sizes := len(data)
	if sizes < 25 {
		data = append(data, bytes.Repeat([]byte{0}, 24-sizes)...)
		return append([]byte{uint8(sizes), uint8(0x00)}, data...)
	}
	num, _ := stackGetNum(sizes)
	poss := stackGetBlocks( uint32(num) )
	dat := make([]byte, 0)
	index := 0
	size := sizes
	STACK.Lock()
	fp, _ := os.OpenFile(ROOT + "stack.s", os.O_WRONLY, 0777)
	for index < num {
		dat = make([]byte, 0)
		for i := 0; i < size; i += 256 {
			fp.Seek( poss[index], os.SEEK_SET )
			if i+256 > size {
				fp.Write( data[i:size] )
			} else {
				fp.Write( data[i : i+256] )
			}
			dat = append(dat, VarsToBytes(poss[index])...)
			index++
		}
		data = dat
		size = len(data)
	}
	fp.Close()
	STACK.Unlock()
	if sizes&0xFF == 0 {
		return VarsToBytes(uint16(0xFF00), poss[len(poss)-1], sizes, sizes)
	}
	return VarsToBytes(uint16(0xFF00), poss[len(poss)-1], sizes, sizes + 256 - (sizes&0xFF))
}

func StackAdd( info, data []byte ) []byte {
	shortLength := info[0]
	if shortLength < 25 {
		lensNew := len(data) + int(shortLength)
		if lensNew<26 {
			info[0] += uint8(len(data))
			copy( info[2+shortLength:2+lensNew], data )
			return info
		}
		return StackNew( append( info[2:2+shortLength], data... ) )
	}
	pos := int(Uint64Big(info[2:10]))
	size := int(Uint64Big(info[10:18]))
	space := int(Uint64Big(info[18:26]))
	sizeNew := size+len(data)
	if sizeNew>space {
		info = stackEnlarge( pos, size, space, sizeNew )
		pos = int(Uint64Big(info[2:10]))
		space = int(Uint64Big(info[18:26]))
		copy(info[10:18], VarsToBytes(sizeNew))
	}else{
		copy( info[10:18], VarsToBytes( sizeNew ) )
	}
	if space==256 && sizeNew <= 256 {
		STACK.Lock()
		fp, err := os.OpenFile(ROOT + "stack.s", os.O_WRONLY, 0777)
		if err==nil {
			fp.Seek( int64(pos+size), os.SEEK_SET )
			fp.Write(data)
		}
		fp.Close()
		STACK.Unlock()
		return info
	}
	blocks, start, end := stackRangeBlock( int64(size), int64(sizeNew), int64(pos), int64(space) )
	num := stackRangeWrite( blocks, data, start, end )
	if num!=len(data){
		return []byte{}
	}
	return info
}

func StackReplace( info, data []byte, from, to int ) int {
	size := int(Uint64Big(info[10:18]))
	if to-from != len(data) || to>size {
		return 0
	}
	pos := int64(Uint64Big(info[2:10]))
	space := int64(Uint64Big(info[18:26]))
	blocks, start, end := stackRangeBlock( int64(from), int64(to), pos, space )
	num := stackRangeWrite( blocks, data, start, end )
	if num!=len(data){
		return 0
	}
	return num
}

func StackGet( info []byte, from, to int ) []byte{
	pos := int64(Uint64Big(info[2:10]))
	space := int64(Uint64Big(info[18:26]))
	blocks, start, _ := stackRangeBlock( int64(from), int64(to), pos, space )
	res := make([]byte, to-from)
	STACK.RLock()
	fp, err := os.OpenFile(ROOT + "stack.s", os.O_RDONLY, 0777)
	if err!=nil {
		return []byte{}
	}
	lens := len(blocks)-1
	index := 256-start
	fp.Seek( blocks[0]+start, os.SEEK_SET )
	fp.Read(res[0:index])
	for i:=1;i<lens;i++ {
		fp.Seek( blocks[i], os.SEEK_SET )
		fp.Read(res[index:index+256])
		index += 256
	}
	fp.Seek( blocks[lens], os.SEEK_SET )
	fp.Read( res[index:to-from] )
	fp.Close()
	STACK.RUnlock()
	return res
}

func StackGetAll(info []byte) []byte {
	shortLength := info[0]
	if shortLength < 25 {
		return info[2 : 2+int(shortLength)]
	}
	pos := int64(Uint64Big(info[2:10]))
	size := int64(Uint64Big(info[10:18]))
	space := int64(Uint64Big(info[18:26]))-1
	var nums, bits, pt int64 = 0x7C00000000000000, 58, 0
	for (nums&(space-1))==0 {
		nums >>= 5
		bits -= 5
	}
	STACK.RLock()
	fp, err := os.OpenFile(ROOT + "stack.s", os.O_RDONLY, 0777)
	lens := 0
	res := make([]byte, size)
	if err==nil {
		blocks := []int64{ pos }
		for nums>0xFF {
			poss := make( []int64, 0 )
			lens = len(blocks)
			pt = ((size-1)&nums) + (1<<bits)
			fp.Seek( blocks[0], os.SEEK_SET )
			if lens==1 {
				bufn := make( []byte, pt>>(bits-3) )
				fp.Read(bufn)
				poss = make( []int64, pt>>bits )
				binary.Read( bytes.NewBuffer(bufn), binary.BigEndian, &poss )
			}else{
				bufn := make( []byte, 256 )
				posn := make([]int64, 32)
				for i:=0; i<lens-1; i++ {
					fp.Seek( blocks[i], os.SEEK_SET )
					fp.Read(bufn)
					binary.Read( bytes.NewBuffer(bufn), binary.BigEndian, &posn )
					poss = append( poss, posn... )
				}
				fp.Seek( blocks[lens-1], os.SEEK_SET )
				bufn = make( []byte, pt>>(bits-3) )
				fp.Read(bufn)
				posn = make( []int64, pt>>8 )
				binary.Read( bytes.NewBuffer(bufn), binary.BigEndian, &posn )
				poss = append( poss, posn... )
			}
			blocks = poss
			nums >>= 5
			bits -= 5
		}
		lens = len(blocks)-1
		for i:=0;i<lens;i++ {
			fp.Seek( blocks[i], os.SEEK_SET )
			fp.Read(res[(i<<8):(i+1)<<8])
		}
		fp.Seek( blocks[lens], os.SEEK_SET )
		fp.Read( res[(lens<<8):size] )
	}
	fp.Close()
	STACK.RUnlock()
	return res
}

func StackStruct(info []byte) []byte {
	pos := int64(Uint64Big(info[2:10]))
	size := int64(Uint64Big(info[10:18]))
	space := int64(Uint64Big(info[18:26]))-1
	var nums, bits, pt int64 = 0x7C00000000000000, 58, 0
	for (nums&(space-1))==0 {
		nums >>= 5
		bits -= 5
	}
	STACK.RLock()
	fp, err := os.OpenFile(ROOT + "stack.s", os.O_RDONLY, 0777)
	defer fp.Close()
	lens := 0
	res := make([]byte, size)
	if err==nil {
		blocks := []int64{ pos }
		for nums>0xFF {
			
			poss := make( []int64, 0 )
			lens = len(blocks)
			pt = ((size-1)&nums) + (1<<bits)
			fmt.Println( "blocks:", blocks)
			fp.Seek( blocks[0], os.SEEK_SET )
			if lens==1 {
				bufn := make( []byte, pt>>(bits-3) )
				fp.Read(bufn)
				poss = make( []int64, pt>>bits )
				binary.Read( bytes.NewBuffer(bufn), binary.BigEndian, &poss )
			}else{
				bufn := make( []byte, 256 )
				posn := make([]int64, 32)
				for i:=0; i<lens-1; i++ {
					fp.Seek( blocks[i], os.SEEK_SET )
					fp.Read(bufn)
					binary.Read( bytes.NewBuffer(bufn), binary.BigEndian, &posn )
					poss = append( poss, posn... )
				}
				fp.Seek( blocks[lens-1], os.SEEK_SET )
				bufn = make( []byte, pt>>(bits-3) )
				fp.Read(bufn)
				posn = make( []int64, pt>>bits )
				
				binary.Read( bytes.NewBuffer(bufn), binary.BigEndian, &posn )
				poss = append( poss, posn... )
			}
			blocks = poss
			nums >>= 5
			bits -= 5
		}
		fmt.Println(blocks)
		lens = len(blocks)-1
		for i:=0;i<lens;i++ {
			fp.Seek( blocks[i], os.SEEK_SET )
			fp.Read(res[(i<<8):(i+1)<<8])
		}
		fp.Seek( blocks[lens], os.SEEK_SET )
		fp.Read( res[(lens<<8):size] )
	}
	STACK.RUnlock()
	return res
}