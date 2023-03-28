package idefixgo

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jaracil/ei"
	be "github.com/nayarsystems/bstates"
	"github.com/nayarsystems/idefix-go/messages"
	"github.com/nayarsystems/idefix/libraries/eval"
)

type GetBstatesParams struct {
	*GetEventsBaseParams
	ForceTsField         string
	RawTsFieldYearOffset uint
	RawTsFieldFactor     float32
}

type Bstate struct {
	Timestamp time.Time
	State     *be.State
	Delta     map[string]any
}

type BstatesSource struct {
	Meta    map[string]interface{}
	MetaRaw string
	States  []*Bstate

	timestampField           string
	timestampFieldYearOffset int
	timestampFieldFactor     float32
}

// domain -> address -> schema -> meta-hash -> source of states
type GetBstatesResult = map[string]map[string]map[string]map[string]*BstatesSource

func GetBstates(ic *Client, p *GetBstatesParams, stateMap GetBstatesResult) (cid string, err error) {
	m, err := ic.GetEventsByDomain(p.Domain, p.Since, p.Limit, p.Cid, p.Timeout)
	if err != nil {
		return
	}
	cid = m.ContinuationID

	for _, e := range m.Events {
		if p.AddressFilter != "" && e.Address != p.AddressFilter {
			continue
		}
		if parseEvent, _ := evalMeta(e.Meta, p.MetaFilter); !parseEvent {
			continue
		}
		schemaId, serr := messages.BstatesParseSchemaIdFromType(e.Type)
		if serr != nil {
			continue
		}
		var blob []byte
		var payloadErr error = nil
		rawMsi, ok := e.Payload.(map[string]interface{})
		if ok {
			blobI, ok := rawMsi["Data"]
			if !ok {
				fmt.Println("no 'Data' field found")
				continue
			}

			switch v := blobI.(type) {
			case []byte:
				blob = v
			case string:
				var derr error
				blob, derr = base64.StdEncoding.DecodeString(v)
				if derr != nil {
					payloadErr = fmt.Errorf("'Data' is a string but is not valid base64: %v", derr)
				}
			default:
				payloadErr = fmt.Errorf("can't get a buffer from 'Data' field")
			}
			if payloadErr != nil {
				fmt.Printf("%v\n", payloadErr)
				continue
			}
		} else {
			b64Str, ok := e.Payload.(string)
			if !ok {
				fmt.Println("wrong payload format")
			}
			var derr error
			blob, derr = base64.StdEncoding.DecodeString(b64Str)
			if derr != nil {
				fmt.Printf("payload is a string but is not a valid base64: %v", derr)
				continue
			}
		}

		var schema *be.StateSchema
		if schema, ok = schemasMap[schemaId]; !ok {
			schemaMsg, serr := ic.GetSchema(schemaId, time.Second)
			if serr != nil {
				fmt.Printf("schema '%s' was not found: %v\n", schemaId, serr)
				continue
			}
			schema = &be.StateSchema{}
			serr = schema.UnmarshalJSON([]byte(schemaMsg.Payload))
			if serr != nil {
				fmt.Printf("can't parse schema '%s': %v\n", schemaId, serr)
				continue
			}
			schemasMap[schemaId] = schema
		}

		var domainMap map[string]map[string]map[string]*BstatesSource
		if domainMap, ok = stateMap[e.Domain]; !ok {
			domainMap = map[string]map[string]map[string]*BstatesSource{}
			stateMap[e.Domain] = domainMap
		}

		var addressMap map[string]map[string]*BstatesSource
		if addressMap, ok = domainMap[e.Address]; !ok {
			addressMap = map[string]map[string]*BstatesSource{}
			domainMap[e.Address] = addressMap
		}

		var schemaMap map[string]*BstatesSource
		if schemaMap, ok = addressMap[schemaId]; !ok {
			schemaMap = map[string]*BstatesSource{}
			addressMap[schemaId] = schemaMap
		}

		metaRaw, merr := json.Marshal(e.Meta)
		if merr != nil {
			fmt.Printf("can't get raw meta: %v\n", merr)
			continue
		}
		metaHashRaw := sha256.Sum256(metaRaw)
		metaHash := base64.StdEncoding.EncodeToString(metaHashRaw[:])

		stateSource, ok := schemaMap[metaHash]
		if !ok {
			stateSource = &BstatesSource{
				Meta:    e.Meta,
				MetaRaw: string(metaRaw),
			}
			schemaMap[metaHash] = stateSource
		}

		// get timestamp field
		dfields := schema.GetDecodedFields()
		tsFieldName := ""
		for _, f := range dfields {
			if f.Decoder.Name() == be.NumberToUnixTsMsDecoderType {
				if p.ForceTsField == "" {
					tsFieldName = f.Name
					break
				} else {
					if p.ForceTsField == f.Name {
						tsFieldName = f.Name
						break
					}
				}
			}
		}
		if tsFieldName == "" && p.ForceTsField != "" {
			// Let's check if forceTsField is a raw numeric field
			rawFields := schema.GetFields()
			for _, f := range rawFields {
				if f.Name == p.ForceTsField {
					tsFieldName = f.Name
					stateSource.timestampFieldYearOffset = int(p.RawTsFieldYearOffset)
					stateSource.timestampFieldFactor = p.RawTsFieldFactor
					break
				}
			}
		}
		stateSource.timestampField = tsFieldName

		var states []*Bstate
		states, err = getStatesList(stateSource, schema, blob)
		if err != nil {
			return
		}
		stateSource.States = append(stateSource.States, states...)
	}
	return
	// for domain, domainMap := range stateMap {
	// 	for address, addressMap := range domainMap {
	// 		for schema, schemaMap := range addressMap {
	// 			for _, statesSource := range schemaMap {
	// 				header := fmt.Sprintf("~~~~~~~~~ DOMAIN: %s, ADDRESS: %s, SCHEMA: %s, META: %s~~~~~~~~~", domain, address, schema, statesSource.MetaRaw)
	// 				headerSeparatorRune := []rune(header)
	// 				for i := 0; i < len(headerSeparatorRune); i++ {
	// 					headerSeparatorRune[i] = '~'
	// 				}
	// 				headerSeparator := string(headerSeparatorRune)
	// 				fmt.Println(headerSeparator)
	// 				fmt.Println(header)
	// 				fmt.Println(headerSeparator)
	// 				events, err := getDeltaStates(statesSource.States)
	// 				if err != nil {
	// 					fmt.Println(err)
	// 					continue
	// 				}
	// 				for i, event := range events {
	// 					je, err := json.Marshal(event)
	// 					if err != nil {
	// 						fmt.Println(err)
	// 						continue
	// 					}
	// 					if statesSource.timestampField != "" {
	// 						ts, err := getTimestampValue(statesSource, i)
	// 						if err != nil {
	// 							fmt.Printf("fix me: %v", err)
	// 							fmt.Println(string(je))
	// 							continue
	// 						}
	// 						fmt.Printf("%v: %s\n", ts, string(je))
	// 					} else {
	// 						fmt.Println(string(je))
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// }
}

func getTimestampValue(s *be.State, tsField string, tsYearOffset int, tsFactor float32) (time.Time, error) {
	tsmsI, err := s.Get(tsField)
	if err != nil {
		return time.Time{}, fmt.Errorf("fix me: cannot get '%s' value: %v", tsField, err)
	}
	tsms, err := ei.N(tsmsI).Float64()
	if err != nil {
		return time.Time{}, fmt.Errorf("fix me: cannot get '%s' value: %v", tsField, err)
	}
	if tsYearOffset == 0 {
		return time.UnixMilli(int64(tsms)), nil
	}
	offsetDate := time.Date(int(tsYearOffset), time.January, 1, 0, 0, 0, 0, time.UTC)
	offsetDateUnixMs := offsetDate.UnixMilli()
	// convert to millis using given factor
	unixTsMs := uint64(offsetDateUnixMs + int64(tsms*float64(tsFactor)))
	return time.UnixMilli(int64(unixTsMs)), nil
}

func evalMeta(meta map[string]interface{}, expr eval.CompiledExpr) (bool, error) {
	metaEnv := map[string]interface{}{}
	for k, v := range meta {
		metaEnv[k] = v
	}
	res := eval.EvalCompiled(expr, metaEnv)
	switch res.Res {
	case eval.ResInvalidExpr, eval.ResInvalidOp, eval.ResTypeMismatch:
		return false, fmt.Errorf("filter expression error (%d): %s", res.Res, res.Iden)
	}
	if res.Res == eval.ResOK {
		return true, nil
	}
	return false, nil
}

func getStatesList(source *BstatesSource, schema *be.StateSchema, raw []byte) (res []*Bstate, err error) {
	decoder := be.CreateStateQueue(schema)
	err = decoder.Decode([]byte(raw))
	if err != nil {
		return nil, fmt.Errorf("can't decode event: %v", err)
	}
	states, err := decoder.GetStates()
	if err != nil {
		return nil, fmt.Errorf("can't decode event: %v", err)
	}
	dstates, err := getDeltaStates(states)
	if err != nil {
		return nil, fmt.Errorf("can't get deltas: %v", err)
	}
	for i, s := range states {
		bstate := &Bstate{
			State: s,
			Delta: dstates[i],
		}
		if source.timestampField != "" {
			ts, err := getTimestampValue(s, source.timestampField, source.timestampFieldYearOffset, source.timestampFieldFactor)
			if err != nil {
				return nil, err
			}
			bstate.Timestamp = ts
		}
		res = append(res, bstate)
	}
	return res, err
}

func getDeltaStates(events []*be.State) ([]map[string]interface{}, error) {
	msiEvents, err := be.GetDeltaMsiStates(events)
	return msiEvents, err
}

var schemasMap = map[string]*be.StateSchema{}
