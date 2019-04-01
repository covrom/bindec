package bindec

import (
	"fmt"
	"go/format"
	"strings"
)

// Options to configure generation.
type Options struct {
	// Path of the package in which the type is located.
	Path string
	// Types to generate encoder and decoder for.
	Types []string
	// Recvs are the receiver names for the generated methods.
	Recvs []string
}

// Generate a file of source code containing an encoder and a decoder to
// encode and decode a given type to and from a binary representation of
// itself.
func Generate(opts Options) ([]byte, error) {
	pkg, err := getPackage(opts.Path)
	if err != nil {
		return nil, err
	}

	ctx := newParseContext()
	ctx.addImport("encoding/binary")
	ctx.addImport("bytes")
	ctx.addImport("io")
	ctx.addImport("math")

	var methods = make([]string, len(opts.Types))
	for i, tName := range opts.Types {
		recv := opts.Recvs[i]

		typ, err := findType(pkg, tName)
		if err != nil {
			return nil, err
		}

		t, err := parseType(ctx, typ)
		if err != nil {
			return nil, err
		}

		methods[i] = generateMethods(recv, tName, t)
	}

	src := []byte(generateFile(
		pkg.Name(),
		strings.Join(methods, "\n"),
		ctx.getImports(),
	))

	formatted, err := format.Source(src)
	if err != nil {
		return nil, fmt.Errorf("error formatting code: %s\n\n%s", err, prettySource(src))
	}

	return formatted, nil
}

func generateMethods(recv, typeName string, typ Type) string {
	return fmt.Sprintf(
		methodsTpl,
		recv,
		typeName,
		typ.Encoder(recv),
		typ.Decoder(recv, true),
	)
}

func generateFile(
	pkgName string,
	methods string,
	imports []string,
) string {
	var deps = make([]string, len(imports))
	for i, x := range imports {
		deps[i] = fmt.Sprintf("%q", x)
	}

	return fmt.Sprintf(
		fileTpl,
		pkgName,
		methods,
		strings.Join(deps, "\n"),
	)
}

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

%[2]s
`

const (
	readString = `
{
	var sz = make([]byte, 8)
	if _, err := io.ReadFull(reader, sz); err != nil {
		return err
	}

	ux := binary.LittleEndian.Uint64(sz)
	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}

	b := make([]byte, int(x))
	if _, err := io.ReadFull(reader, b); err != nil {
		return err
	}

	%[3]s%[2]s = %[1]s(b)
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
	var sz = make([]byte, 8)
	if _, err := io.ReadFull(reader, sz); err != nil {
		return err
	}

	ux := binary.LittleEndian.Uint64(sz)
	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}

	b := make([]byte, int(x))
	if _, err := io.ReadFull(reader, b); err != nil {
		return err
	}

	%[3]s%[1]s = %[2]s(b)
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

func prettySource(src []byte) string {
	lines := strings.Split(string(src), "\n")
	maxDigits := len(fmt.Sprint(len(lines)))
	format := fmt.Sprintf("%%%dd | %%s", maxDigits)

	for i, line := range lines {
		lines[i] = fmt.Sprintf(format, i+1, line)
	}

	return strings.Join(lines, "\n")
}
