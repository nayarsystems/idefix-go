package normalize

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/jaracil/ei"
	"github.com/nayarsystems/buffer/buffer"
	"github.com/nayarsystems/buffer/shuffling"
)

type msi = map[string]interface{}

func MsiDup(data msi) msi {
	resp := msi{}
	for key, val := range data {
		if v, ok := val.(msi); ok {
			resp[key] = MsiDup(v)
		} else {
			resp[key] = val
		}
	}
	return resp
}

func DecodeTypes(data msi) error {
	for key, val := range data {
		if v, ok := val.(msi); ok {
			if err := DecodeTypes(v); err != nil {
				return err
			}
		}
		tokens := strings.Split(key, ":")
		if len(tokens) < 2 {
			continue
		}
		targetData := val
		for i := len(tokens) - 1; i >= 1; i-- {
			encoding := strings.ToLower(tokens[i])
			if strings.HasPrefix(encoding, "trans(") {
				//format: trans\([0-9]+,[0-9]+\)
				argsRaw := encoding[6 : len(encoding)-1]
				args := strings.Split(argsRaw, ",")
				if len(args) != 2 {
					return fmt.Errorf("trans decoder invalid number of arguments")
				}
				numRows, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("trans decoder num. rows not an integer")
				}
				numCols, err := strconv.Atoi(args[1])
				if err != nil {
					return fmt.Errorf("trans decoder num. cols not an integer")
				}
				if s, ok := targetData.([]byte); ok {
					b := &buffer.Buffer{}
					err := b.InitFromRawBufferN(s, numRows*numCols)
					if err != nil {
						return fmt.Errorf("trans decoder error: %s", err.Error())
					}
					sb, err := shuffling.TransposeBits(b, numCols)
					if err != nil {
						return fmt.Errorf("trans decoder error: %s", err.Error())
					}
					targetData = sb.GetRawBuffer()
				} else {
					return fmt.Errorf("input of trans decoder must be []byte")
				}
				continue
			}
			switch encoding {
			case "dur":
				s, err := ei.N(targetData).String()
				if err != nil {
					return fmt.Errorf("input of duration decoder must be string")
				}
				targetData, err = time.ParseDuration(s)
				if err != nil {
					return fmt.Errorf("input of duration decoder has wrong string format")
				}
			case "time":
				s, err := ei.N(targetData).Int64()
				if err != nil {
					return fmt.Errorf("input of time decoder must be numeric")
				}
				targetData = time.UnixMilli(s)
			case "b64":
				if s, ok := targetData.(string); ok {
					b, err := base64.StdEncoding.DecodeString(s)
					if err != nil {
						return err
					}
					targetData = b
				}
			case "hex":
				if s, ok := targetData.(string); ok {
					b, err := hex.DecodeString(s)
					if err != nil {
						return err
					}
					targetData = b
				}
			case "bytes":
				if s, ok := targetData.(string); ok {
					targetData = []byte(s)
				} else {
					return fmt.Errorf("input of bytes decoder must be string")
				}
			case "string":
				if s, ok := targetData.([]byte); ok {
					targetData = string(s)
				} else {
					return fmt.Errorf("input of string decoder must be []byte")
				}
			case "gzip":
				if s, ok := targetData.([]byte); ok {
					r := bytes.NewReader(s)
					gzr, err := gzip.NewReader(r)
					if err != nil {
						return err
					}
					targetData, err = io.ReadAll(gzr)
					gzr.Close()
					if err != nil {
						return err
					}
				} else {
					return fmt.Errorf("input of gzip decoder must be []byte")
				}
			default:
				return fmt.Errorf("invalid encoding %s", encoding)
			}
		}
		data[tokens[0]] = targetData
		delete(data, key)
	}
	return nil
}

type EncodeTypesOpts struct {
	BytesToB64    bool // Encode []byte fields to base64 enable when JSON encoding needed.
	Compress      bool // Autocompress []byte fields when size is greater or equal than CompThreshold
	CompThreshold int  // Threshold in bytes-length for autocompression
}

func EncodeTypes(data msi, opts *EncodeTypesOpts) error {
	for key, val := range data {
		if v, ok := val.(msi); ok {
			EncodeTypes(v, opts)
		}

		if strings.Contains(key, ":") {
			continue
		}

		targetKey := key
		targetData := val

		switch v := targetData.(type) {
		case []byte:
			if opts.Compress && len(v) >= opts.CompThreshold {
				buf := new(bytes.Buffer)
				wr := gzip.NewWriter(buf)
				n, err := wr.Write(v)
				wr.Close()
				if err == nil && n == len(v) {
					comp, err := io.ReadAll(buf)
					if err == nil && len(comp) < len(v) {
						targetKey = targetKey + ":gzip"
						targetData = comp
					}
				}
			}
			if opts.BytesToB64 {
				targetKey = targetKey + ":b64"
				targetData = base64.StdEncoding.EncodeToString(targetData.([]byte))
			}
		case time.Duration:
			targetKey = targetKey + ":dur"
			targetData = v.String()
		case time.Time:
			targetKey = targetKey + ":time"
			targetData = v.UnixMilli()
		}
		if targetKey != key {
			data[targetKey] = targetData
			delete(data, key)
		}
	}

	return nil
}
