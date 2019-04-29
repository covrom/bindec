package bindec

const methodsTpl = `
// EncodeBinary returns a binary-encoded representation of the type.
func (%[1]s %[2]s) EncodeBinary() ([]byte, error) {
	var writer = bytes.NewBuffer(nil)
	if err := %[1]s.WriteBinary(writer); err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

// WriteBinary writes the binary-encoded representation of the type to the
// given writer.
func (%[1]s %[2]s) WriteBinary(writer io.Writer) error {
	%[3]s
	return nil
}

// DecodeBinaryFromBytes fills the type with the given binary-encoded
// representation of the type.
func (%[1]s *%[2]s) DecodeBinaryFromBytes(data []byte) error {
	var reader = bytes.NewReader(data)
	return %[1]s.DecodeBinary(reader)
}

// DecodeBinary reads the binary representation of the type from the given
// reader and fulls the type with it.
func (%[1]s *%[2]s) DecodeBinary(reader io.Reader) error {
	%[4]s
	return nil
}
`

const fileTpl = `
// WARNING! This is code generated by bindec, do not modify manually.

package %[1]s

import (
	%[3]s
)

var _ = binary.LittleEndian
var _ = math.Abs

%[4]s

%[2]s
`

const (
	readString = `
{
	var bs = make([]byte, 8)
	if _, err := io.ReadFull(reader, bs); err != nil {
		return err
	}

	ux := binary.LittleEndian.Uint64(bs)
	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}

	sz := int(x)
	%[4]s

	b := make([]byte, sz)
	if _, err := io.ReadFull(reader, b); err != nil {
		return err
	}

	%[3]s%[2]s = %[1]s(b)

	%[5]s
}
`

	writeString = `
{
	v := %s
	len := len(v)
	ux := uint64(len) << 1
	if len < 0 {
		ux = ^ux
	}
	sz := make([]byte, 8)
	binary.LittleEndian.PutUint64(sz, ux)
	if _, err := writer.Write(sz); err != nil {
		return err
	}

	_, err := writer.Write([]byte(v))
	if err != nil {
		return err
	}
}
`

	readBool = `
{
	var v = make([]byte, 1)
	if _, err := io.ReadFull(reader, v); err != nil {
		return err
	}

	%[3]s%[2]s = %[1]s(v[0] == 1)

	%[4]s
	%[5]s
}
`

	writeBool = `
{
	var v byte
	if %s {
		v = 1
	}
	_, err := writer.Write([]byte{v})
	if err != nil {
		return err
	}
}
`

	readInt = `
{
	var bs = make([]byte, 8)
	if _, err := io.ReadFull(reader, bs); err != nil {
		return err
	}

	ux := binary.LittleEndian.Uint64(bs)
	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}
	%[3]s%[2]s = %[1]s(x)

	%[4]s
	%[5]s
}
`

	writeInt = writeInt64

	readUint = `
{
	var bs = make([]byte, 8)
	if _, err := io.ReadFull(reader, bs); err != nil {
		return err
	}

	ux := binary.LittleEndian.Uint64(bs)
	%[3]s%[2]s = %[1]s(ux)

	%[4]s
	%[5]s
}
`

	writeUint = `
	{
		x := %s
		bs := make([]byte, 8)
		binary.LittleEndian.PutUint64(bs, uint64(x))
		_, err := writer.Write(bs)
		if err != nil {
			return err
		}
	}
`

	readInt64 = `
{
	var bs = make([]byte, 8)
	if _, err := io.ReadFull(reader, bs); err != nil {
		return err
	}

	ux := binary.LittleEndian.Uint64(bs)
	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}
	%[3]s%[2]s = %[1]s(x)

	%[4]s
	%[5]s
}
`

	writeInt64 = `
{
	x := %s
	ux := uint64(x) << 1
	if x < 0 {
		ux = ^ux
	}
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, ux)
	_, err := writer.Write(bs)
	if err != nil {
		return err
	}
}
`

	readUintptr = `
{
	var bs = make([]byte, 8)
	if _, err := io.ReadFull(reader, bs); err != nil {
		return err
	}

	ux := binary.LittleEndian.Uint64(bs)
	%[3]s%[2]s = %[1]s(ux)

	%[4]s
	%[5]s
}
`

	writeUintptr = `
{
	x := uint64(%s)
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, x)
	_, err := writer.Write(bs)
	if err != nil {
		return err
	}
}
`

	readUint64 = `
{
	var bs = make([]byte, 8)
	if _, err := io.ReadFull(reader, bs); err != nil {
		return err
	}

	ux := binary.LittleEndian.Uint64(bs)
	%[3]s%[2]s = %[1]s(ux)

	%[4]s
	%[5]s
}
`

	writeUint64 = `
{
	x := uint64(%s)
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, x)
	_, err := writer.Write(bs)
	if err != nil {
		return err
	}
}
`

	readInt32 = `
{
	var bs = make([]byte, 4)
	if _, err := io.ReadFull(reader, bs); err != nil {
		return err
	}

	ux := binary.LittleEndian.Uint32(bs)
	x := int32(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}
	%[3]s%[2]s = %[1]s(x)

	%[4]s
	%[5]s
}
`

	writeInt32 = `
{
	x := int32(%s)
	ux := uint32(x) << 1
	if x < 0 {
		ux = ^ux
	}
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, ux)
	_, err := writer.Write(bs)
	if err != nil {
		return err
	}
}
`

	readUint32 = `
{
	var bs = make([]byte, 4)
	if _, err := io.ReadFull(reader, bs); err != nil {
		return err
	}

	ux := binary.LittleEndian.Uint32(bs)
	%[3]s%[2]s = %[1]s(ux)

	%[4]s
	%[5]s
}
`

	writeUint32 = `
{
	x := uint32(%s)
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, x)
	_, err := writer.Write(bs)
	if err != nil {
		return err
	}
}
`

	readInt16 = `
{
	var bs = make([]byte, 2)
	if _, err := io.ReadFull(reader, bs); err != nil {
		return err
	}

	ux := binary.LittleEndian.Uint16(bs)
	x := int16(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}
	%[3]s%[2]s = %[1]s(x)

	%[4]s
	%[5]s
}
`

	writeInt16 = `
{
	x := int16(%s)
	ux := uint16(x) << 1
	if x < 0 {
		ux = ^ux
	}
	bs := make([]byte, 2)
	binary.LittleEndian.PutUint16(bs, ux)
	_, err := writer.Write(bs)
	if err != nil {
		return err
	}
}
`

	readUint16 = `
{
	var bs = make([]byte, 2)
	if _, err := io.ReadFull(reader, bs); err != nil {
		return err
	}

	ux := binary.LittleEndian.Uint16(bs)
	%[3]s%[2]s = %[1]s(ux)

	%[4]s
	%[5]s
}
`

	writeUint16 = `
{
	x := uint16(%s)
	bs := make([]byte, 2)
	binary.LittleEndian.PutUint16(bs, x)
	_, err := writer.Write(bs)
	if err != nil {
		return err
	}
}
`

	readInt8 = `
{
	var bs = make([]byte, 1)
	if _, err := io.ReadFull(reader, bs); err != nil {
		return err
	}

	ux := bs[0]
	x := int8(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}
	%[3]s%[2]s = %[1]s(x)

	%[4]s
	%[5]s
}
`

	writeInt8 = `
{
	x := int8(%s)
	ux := byte(x) << 1
	if x < 0 {
		ux = ^ux
	}
	_, err := writer.Write([]byte{ux})
	if err != nil {
		return err
	}
}
`

	readByte = `
{
	var bs = make([]byte, 1)
	if _, err := io.ReadFull(reader, bs); err != nil {
		return err
	}
	%[3]s%[2]s = %[1]s(bs[0])

	%[4]s
	%[5]s
}
`

	writeByte = `
{
	if _, err := writer.Write([]byte{byte(%s)}); err != nil {
		return err
	}
}
`

	readFloat32 = `
{
	var bs = make([]byte, 4)
	if _, err := io.ReadFull(reader, bs); err != nil {
		return err
	}
	ux := binary.LittleEndian.Uint32(bs)
	%[3]s%[2]s = %[1]s(math.Float32frombits(ux))

	%[4]s
	%[5]s
}
`

	writeFloat32 = `
{
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, math.Float32bits(float32(%s)))
	_, err := writer.Write(bs)
	if err != nil {
		return err
	}
}
`

	readFloat64 = `
{
	var bs = make([]byte, 8)
	if _, err := io.ReadFull(reader, bs); err != nil {
		return err
	}
	ux := binary.LittleEndian.Uint64(bs)
	%[3]s%[2]s = %[1]s(math.Float64frombits(ux))

	%[4]s
	%[5]s
}
`

	writeFloat64 = `
{
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, math.Float64bits(float64(%s)))
	_, err := writer.Write(bs)
	if err != nil {
		return err
	}
}
`

	readBytes = `
{
	var bs = make([]byte, 8)
	if _, err := io.ReadFull(reader, bs); err != nil {
		return err
	}

	ux := binary.LittleEndian.Uint64(bs)
	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}

	sz := int(x)
	%[4]s

	b := make([]byte, sz)
	if _, err := io.ReadFull(reader, b); err != nil {
		return err
	}

	%[3]s%[1]s = %[2]s(b)

	%[5]s
}
`

	writeBytes = `
{
	v := %s
	len := len(v)
	ux := uint64(len) << 1
	if len < 0 {
		ux = ^ux
	}
	sz := make([]byte, 8)
	binary.LittleEndian.PutUint64(sz, ux)
	if _, err := writer.Write(sz); err != nil {
		return err
	}

	_, err := writer.Write([]byte(v))
	if err != nil {
		return err
	}
}
`
)
