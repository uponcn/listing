package listing

import(
	"encoding/binary"
	"bytes"
)

func NumToBytes(num int) []byte {
	if num <= 0 {
		return []byte{0}
	}
	rs := []byte{}
	for num > 0 {
		rs = append(rs, uint8(num&0x7F)|0x80)
		num = num >> 7
	}
	rs[0] = rs[0] & 0x7F
	n := len(rs) >> 1
	for i := 0; i < n; i++ {
		rs[i], rs[len(rs)-1-i] = rs[len(rs)-1-i], rs[i]
	}
	return rs
}

func NumsToBytes(nums []int) []byte {
	res := make([]byte, 0)
	for _, num := range nums {
		if num <= 0 {
			continue
		}
		rs := []byte{}
		for num > 0 {
			rs = append(rs, uint8(num&0x7F)|0x80)
			num = num >> 7
		}
		rs[0] = rs[0] & 0x7F
		n := len(rs) >> 1
		for i := 0; i < n; i++ {
			rs[i], rs[len(rs)-1-i] = rs[len(rs)-1-i], rs[i]
		}
		res = append(res, rs...)
	}
	return res
}

func VarsToBytes( args ...interface{} ) []byte {
	buf := new(bytes.Buffer)
	for _, arg := range args {
		switch arg.(type) {
		case string:
			binary.Write(buf, binary.BigEndian, NumToBytes(len(arg.(string))))
            binary.Write(buf, binary.BigEndian, []byte(arg.(string)))
		case int:
			binary.Write(buf, binary.BigEndian, int64(arg.(int)))
		case uint:
			binary.Write(buf, binary.BigEndian, uint64(arg.(uint)))
		case []int:
			for _,n := range arg.([]int) {
				binary.Write(buf, binary.BigEndian, int64(n))
			}
		case []uint:
			for _,n := range arg.([]uint) {
				binary.Write(buf, binary.BigEndian, uint64(n))
			}
		// int8,uint8,int16,uint16,int32,uint32,int64,uint64,float32,float64,bool,
		//[]byte,[]int8,[]int16,[]uint16,[]int32,[]uint32,[]int64,[]uint64,[]float32,[]float64,[]bool:
		default:
			binary.Write(buf, binary.BigEndian, arg)
		}
	}
	return buf.Bytes()
}

func Scan( data []byte, args ...interface{}) int {
	buf := bytes.NewBuffer(data)
    var by uint8
	for _, arg := range args {
		switch arg.(type) {
		case *string:
            lens := 0
			for{
                binary.Read(buf, binary.BigEndian, &by)
                lens = (lens<<7) | int(by&0x7F)
                if by<0x80 {
                    break
                }
            }
            bys := make( []byte, lens )
            binary.Read(buf, binary.BigEndian, &bys)
            *arg.(*string) = string(bys)
		case *int:
			var n int64
			binary.Read(buf, binary.BigEndian, &n)
			*arg.(*int) = int(n)
		case *uint:
			var n uint64
			binary.Read(buf, binary.BigEndian, &n)
			*arg.(*uint) = uint(n)
		default:
			binary.Read(buf, binary.BigEndian, arg)
		}
	}
	return 0
}

func Uint16Big(by []byte) uint16 {
	return binary.BigEndian.Uint16(by)
}

func Uint32Big(by []byte) uint32 {
	return binary.BigEndian.Uint32(by)
}

func Uint64Big(by []byte) uint64 {
	return binary.BigEndian.Uint64(by)
}

func Uint16Little(by []byte) uint16 {
	return binary.LittleEndian.Uint16(by)
}

func Uint32Little(by []byte) uint32 {
	return binary.LittleEndian.Uint32(by)
}

func Uint64Little(by []byte) uint64 {
	return binary.LittleEndian.Uint64(by)
}