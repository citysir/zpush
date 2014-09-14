package json

import (
	"bufio"
	"bytes"
	"io/ioutil"
)

func FormatBytes(bytes []byte) ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := bufio.NewWriter(buffer)
	for _, c := range bytes {
		switch c {
		case '\\':
			writer.WriteString("\\\\")
		case '/':
			writer.WriteString("\\/")
		case '"':
			writer.WriteString("\\\"")
		case '\t':
			writer.WriteString("\\t")
		case '\f':
			writer.WriteString("\\f")
		case '\b':
			writer.WriteString("\\b")
		case '\n':
			writer.WriteString("\\n")
		case '\r':
			writer.WriteString("\\r")
		default:
			err := writer.WriteByte(c)
			if err != nil {
				return nil, err
			}
		}
	}
	writer.Flush()
	return ioutil.ReadAll(out)
}

func FormatString(str string) (string, error) {
	bytes, err := FormatBytes([]byte(str))
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
