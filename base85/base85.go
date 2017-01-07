package base85

import (
	//"io"
	"strconv"
)

//0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz!#$%&()*+-;<=>?@^_`{|}~
var b2c_table = [85]byte{
	// 0     1     2     3     4     5     6     7     8     9     A     B     C     D     E     F
	0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46,//0
	0x47, 0x48, 0x49, 0x4A, 0x4B, 0x4C, 0x4D, 0x4E, 0x4F, 0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56,//1
	0x57, 0x58, 0x59, 0x5A, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6A, 0x6B, 0x6C,//2
	0x6D, 0x6E, 0x6F, 0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x78, 0x79, 0x7A, 0x21, 0x23,//3
	0x24, 0x25, 0x26, 0x28, 0x29, 0x2A, 0x2B, 0x2D, 0x3B, 0x3C, 0x3D, 0x3E, 0x3F, 0x40, 0x5E, 0x5F,//4
	0x60, 0x7B, 0x7C, 0x7D, 0x7E,
}

var encode [85]byte
var decode [256]byte

func init() {
	copy(encode[:], b2c_table[:])
	for i := 0; i < len(decode); i++ {
		decode[i] = 0xFF
	}
	for i := 0; i < len(encode); i++ {
		decode[encode[i]] = byte(i)
	}
}

//encodeChunk encode 4 byte-chunk to 5 byte
//if chunk size is less then 4, then it is padded before convertion.
func encodeChunk(dst, src []byte) int {
	if len(src) == 0 {
		return 0
	}

	//read 4 byte as big-endian uint32 into small endian uint32
	var val uint32
	switch len(src) {
	default:
		val |= uint32(src[3])
		fallthrough
	case 3:
		val |= uint32(src[2]) << 8
		fallthrough
	case 2:
		val |= uint32(src[1]) << 16
		fallthrough
	case 1:
		val |= uint32(src[0]) << 24
	}


	buf := [5]byte{0,0,0,0,0}

	for i := 4; i >= 0; i-- {
		r := val % 85
		val /= 85
		buf[i] = encode[r]
	}

	m := EncodedLen(len(src))
	copy(dst[:], buf[:m])
	return m
}

var decode_base = [5]uint32{85*85*85*85, 85*85*85, 85*85, 85, 1}
//encodeChunk encode 5 byte-chunk to 4 byte
//if chunk size is less then 5, then it is padded before convertion.
func decodeChunk(dst, src []byte) (int, error) {
	if len(src) == 0 {
		return 0, nil
	}
	var val uint32
	m := DecodedLen(len(src))
	buf := [5]byte{84, 84, 84, 84, 84}
	for i := 0; i < len(src); i++ {
		e := decode[src[i]]
		if e == 0xFF {
			return 0, CorruptInputError(i)
		}
		buf[i] = e
	}

	for i := 0; i < 5; i++ {
		r := buf[i]
		val += uint32(r) * decode_base[i]
	}
	//small endian uint32 to big endian uint32 in bytes
	switch m {
	default:
		dst[3] = byte(val & 0xff)
		fallthrough
	case 3:
		dst[2] = byte((val >> 8) & 0xff)
		fallthrough
	case 2:
		dst[1] = byte((val >> 16) & 0xff)
		fallthrough
	case 1:
		dst[0] = byte((val >> 24) & 0xff)
	}
	return m, nil
}

//Encode encodes src into dst, return the bytes written
func Encode(dst, src []byte) int {
	n := 0
	for len(src) > 0{
		if len(src) < 4 {
			n += encodeChunk(dst, src)
			return n
		}
		n += encodeChunk(dst[:5], src[:4])
		src = src[4:]
		dst = dst[5:]
	}
	return n
}

func Decode(dst, src []byte) (int, error) {
	f := 0
	t := 0
	for len(src) > 0{
		if len(src) < 5 {
			w, err := decodeChunk(dst, src)
			return t+w, err
		}

		_, err := decodeChunk(dst[:4], src[:5])
		if err != nil {
			return f ,err
		} else {
			t += 4
			f += 5
			src = src[5:]
			dst = dst[4:]
		}
	}
	return f, nil
}


func EncodedLen(n int) int {
	s := n / 4
	r := n % 4
	if r > 0 {
		return s * 5 + 5 - (4 - r)
	} else {
		return s * 5
	}
}

func DecodedLen(n int) int {
	s := n / 5
	r := n % 5
	if r > 0 {
		return s * 4 + 4 - (5 - r)
	} else {
		return s * 4
	}
}
type CorruptInputError int64

func (e CorruptInputError) Error() string {
	return "illegal ascii85 data at input byte " + strconv.FormatInt(int64(e), 10)
}
