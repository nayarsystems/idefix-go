package idefixgo

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jaracil/ei"
	be "github.com/nayarsystems/bstates"
	"github.com/nayarsystems/idefix-go/eval"
	"github.com/nayarsystems/idefix-go/messages"
)

// GetBstatesParams defines the parameters used to retrieve a list of Bstates from the remote system.
// Each field corresponds to a specific filter or setting for the request, allowing for flexible
// querying of Bstates based on various criteria.
type GetBstatesParams struct {
	UID                  string            // Unique identifier for the event.
	Domain               string            // Specifies the domain from which to retrieve.
	Since                time.Time         // A timestamp indicating the starting point for retrieving Bstates; only Bstates after this time will be returned.
	Limit                uint              // The maximum number of results to return in the response.
	Cid                  string            // ContinuationID lets you get following results after your last request.
	AddressFilter        string            // An optional filter to limit results based on specific addresses.
	MetaFilter           eval.CompiledExpr // A compiled expression used to filter results based on metadata criteria.
	Timeout              time.Duration     // The maximum duration to wait for the request to complete before timing out.
	ForceTsField         string            // Specifies the field used for the timestamp if needed, overriding default behavior.
	RawTsFieldYearOffset uint              // An optional offset to apply to the year field in raw timestamp data.
	RawTsFieldFactor     float32           // A scaling factor applied to the raw timestamp data for accurate representation.
}

// Bstate represents a single snapshot of the state of a device at a specific point in time.
// It consists of a timestamp indicating when the snapshot was recorded and the actual state data.
// For more information see Bstates documentation: https://github.com/nayarsystems/bstates
type Bstate struct {
	Timestamp time.Time
	State     *be.State
}

// BstatesBlob represents a collection of Bstates along with associated metadata.
// It contains information about the unique identifier of the user, the timestamp
// when the collection was created, the actual states, and the raw byte data
// representing the compressed queue of states.
type BstatesBlob struct {
	UID       string
	Timestamp time.Time
	States    []*Bstate
	Raw       []byte
}

// BstatesSource represents a source of Bstates that includes metadata and a collection of
// BstatesBlob instances. It provides contextual information for processing the Bstates.
type BstatesSource struct {
	Meta    map[string]interface{}
	MetaRaw string
	Blobs   []*BstatesBlob

	timestampField           string
	timestampFieldYearOffset int
	timestampFieldFactor     float32
}

// domain -> address -> schema -> meta-hash -> source of states

// GetBstatesResult represents the result of retrieving bstate information
// from the Idefix system. It is structured as a nested map to provide
// easy access to bstate sources.
type GetBstatesResult = map[string]map[string]map[string]map[string]*BstatesSource

// GetSchema retrieves the schema associated with the provided hash from the Idefix service.
//
// This method sends a request to the Idefix service to fetch the schema by hash.
// It uses a message structure that specifies the hash and whether to perform a check.
//
// If the request is successful, the function returns the response containing the schema information.
// If an error occurs during the request, it returns nil along with the error.
func (c *Client) GetSchema(hash string, timeout time.Duration) (*messages.SchemaGetResponseMsg, error) {
	msg := messages.SchemaGetMsg{
		Hash:  hash,
		Check: false,
	}
	resp := &messages.SchemaGetResponseMsg{}
	err := c.Call2("idefix", &messages.Message{To: "schemas.get", Data: &msg}, resp, timeout)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetBstates retrieves bstates events from cloud based on the provided parameters.
//
// If a UID is specified in the parameters, the function first fetches the event associated with that UID.
// It then populates the provided state map with the states extracted from the event. If no states are found,
// an error is returned indicating that the event is not based on bstates.
//
// If no UID is specified, the function retrieves bstates using the provided parameters and populates the state map accordingly.
//
// The function returns the total number of blobs processed, the continuation ID, and any error encountered during the process.
func GetBstates(ic *Client, p *GetBstatesParams, stateMap GetBstatesResult) (totalBlobs uint, cid string, err error) {
	if p.UID != "" {
		var res *messages.EventsGetUIDResponseMsg
		res, err = ic.GetEventByUID(p.UID, p.Timeout)
		if err != nil {
			return
		}
		input := []*messages.Event{
			&res.Event,
		}
		totalBlobs, err = fillStateMap(ic, input, p, stateMap)
		if totalBlobs == 0 {
			err = fmt.Errorf("not a bstates based event")
		}
		return
	}
	totalBlobs, p.Cid, err = getBstates(ic, p, stateMap)
	cid = p.Cid
	return
}

func getBstates(ic *Client, p *GetBstatesParams, stateMap GetBstatesResult) (numblobs uint, cid string, err error) {
	m, err := ic.GetEvents(p.Domain, p.AddressFilter, p.Since, p.Limit, p.Cid, p.Timeout)
	if err != nil {
		return
	}
	cid = m.ContinuationID
	numblobs, err = fillStateMap(ic, m.Events, p, stateMap)
	return
}

func fillStateMap(ic *Client, events []*messages.Event, p *GetBstatesParams, stateMap GetBstatesResult) (numblobs uint, err error) {
	for _, e := range events {
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
		if schema = getSchemaFromCache(schemaId); schema == nil {
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
			saveSchemaOnCache(schemaId, schema)
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
		var beBlob *BstatesBlob
		beBlob, err = createBlobInfo(stateSource, schema, e.UID, e.Timestamp, blob)
		if err != nil {
			return
		}
		stateSource.Blobs = append(stateSource.Blobs, beBlob)
		numblobs += 1
		if numblobs == p.Limit {
			return
		}
	}
	return
}

// BenchmarkBstates measures and compares the sizes of BstatesBlob against different compression pipelines.
//
// This function takes a BstatesBlob and a list of Bstates, extracts the state data, and calculates the uncompressed size.
// It then compares this size with the compressed sizes produced by two different pipelines ("gzip" and "transposed + gzip").
// The results are printed, including the number of states, the uncompressed size, the raw blob size, and the compressed blob sizes.
// If there are no states to process, the function returns immediately.
func BenchmarkBstates(blob *BstatesBlob, bstates []*Bstate) {
	states := []*be.State{}
	for _, s := range bstates {
		states = append(states, s.State)
	}
	if len(states) == 0 {
		return
	}
	stateSize := states[0].GetByteSize()
	fmt.Printf("states count: %d\n", len(states))
	fmt.Printf("state size (B): %d\n", stateSize)
	uncompressedSize := stateSize * len(states)
	fmt.Printf("total states size (B): %d\n", uncompressedSize)
	fmt.Printf("received blob size (B): %d (%.2f %%)\n", len(blob.Raw), float32(len(blob.Raw))/float32(uncompressedSize)*100)

	pipeline := "z"
	size, err := getSizeUsingNewPipeline(states, pipeline)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("blob size using pipeline \"%s\": %d (%.2f %%)\n", pipeline, size, float32(size)/float32(uncompressedSize)*100)

	pipeline = "t:z"
	size, err = getSizeUsingNewPipeline(states, pipeline)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("blob size using pipeline \"%s\": %d (%.2f %%)\n", pipeline, size, float32(size)/float32(uncompressedSize)*100)

}

func getSizeUsingNewPipeline(states []*be.State, pipeline string) (size uint, err error) {
	schema, err := updateSchemaPipeline(states[0].GetSchema(), pipeline)
	if err != nil {
		return
	}
	// Every input state has a reference to the original schema that was used to encode it.
	// The same schema is used by the StateQueue since it needs to know the fields structure of every state
	// to generate a blob (multiple states encoded). The schema contains the pipeline used to encode the StateQueue.
	// We cannot use the input states to create a new StateQueue (blob) that uses the schema with the pipeline updated
	// since it would cause an error due to a schema hash mismatch between the input state's schemas and the new StateQueue
	// (despite having the same field structure).
	// TODO: bstates lib should allow the updates of a StateQueue using a state that has
	// compatible fields structure.
	// So, we need to create new states with the schema with the pipeline updated
	stateQueue := be.CreateStateQueue(schema)
	for _, s := range states {
		rawState, err := s.Encode()
		if err != nil {
			return 0, err
		}
		newState, err := be.CreateState(schema)
		if err != nil {
			return 0, err
		}
		newState.Decode(rawState)
		err = stateQueue.Push(newState)
		if err != nil {
			return 0, err
		}
	}
	blob, err := stateQueue.Encode()
	if err != nil {
		return 0, err
	}
	return uint(len(blob)), nil
}

func updateSchemaPipeline(schema *be.StateSchema, pipeline string) (newSchema *be.StateSchema, err error) {
	schemaMsi := schema.ToMsi()
	schemaMsi["encoderPipeline"] = pipeline
	var newSchemaRaw []byte
	newSchemaRaw, err = json.Marshal(schemaMsi)
	if err != nil {
		return nil, fmt.Errorf("fix me: can't create new schema: %v", err)
	}
	newSchema = &be.StateSchema{}
	err = json.Unmarshal(newSchemaRaw, &newSchema)
	if err != nil {
		return nil, fmt.Errorf("fix me: can't decode new schema: %v", err)
	}
	return
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

// GetDeltaStates processes a slice of Bstate objects and retrieves the delta states
// represented as a slice of maps. Each Bstate contains a State that is processed to
// derive the delta states, which are useful for tracking changes between different
// states over time.
func GetDeltaStates(in []*Bstate) ([]map[string]interface{}, error) {
	var states []*be.State
	for _, bs := range in {
		states = append(states, bs.State)
	}
	msiEvents, err := be.GetDeltaMsiStates(states)
	return msiEvents, err
}

func createBlobInfo(source *BstatesSource, schema *be.StateSchema, uid string, ts time.Time, raw []byte) (res *BstatesBlob, err error) {
	decoder := be.CreateStateQueue(schema)
	err = decoder.Decode([]byte(raw))
	if err != nil {
		return nil, fmt.Errorf("can't decode event: %v", err)
	}
	states, err := decoder.GetStates()
	if err != nil {
		return nil, fmt.Errorf("can't decode event: %v", err)
	}
	res = &BstatesBlob{
		Timestamp: ts,
		UID:       uid,
		Raw:       raw,
	}
	for _, s := range states {
		bstate := &Bstate{
			State: s,
		}
		if source.timestampField != "" {
			ts, err := getTimestampValue(s, source.timestampField, source.timestampFieldYearOffset, source.timestampFieldFactor)
			if err != nil {
				return nil, err
			}
			bstate.Timestamp = ts
		}
		res.States = append(res.States, bstate)
	}
	return res, err
}

var schemasMap = map[string]*be.StateSchema{}
var schemasMapMutex sync.Mutex

func getSchemaFromCache(schemaId string) *be.StateSchema {
	schemasMapMutex.Lock()
	defer schemasMapMutex.Unlock()
	return schemasMap[schemaId]
}

func saveSchemaOnCache(schemaId string, schema *be.StateSchema) {
	schemasMapMutex.Lock()
	defer schemasMapMutex.Unlock()
	schemasMap[schemaId] = schema
}
