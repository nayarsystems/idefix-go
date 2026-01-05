# Bstates Events Example

This example demonstrates how to consume and process bstates events from Idefix.

## What are bstates events?

Bstates events contain batches of states encoded using the [bstates library](https://github.com/nayarsystems/bstates). A state is a key-value snapshot taken at a specific point in time. The encoding schema defines which fields exist, their types, and how they're encoded in binary format.

Events of this type have a mimetype following this pattern:
```
application/vnd.nayar.bstates; id="oEM5eJzBBGbyT9CLrSKrQwdnP2C+CVM8JHjfA0g3MAB="
```

The `id` parameter contains the hash of the schema used to encode the state batch.

## How it works

1. **Connect** to Idefix using user credentials
2. **Request** events filtered by type `application/vnd.nayar.bstates`
3. **Extract** the schema ID from each event's mimetype
4. **Fetch** the schema from the server (with local caching)
5. **Decode** the binary payload using the schema
6. **Process** each state in the batch (this example prints them as JSON)

## Usage

```bash
go run . -user <user_address> -token <user_token> [-a <device_address>] [-d <domain>] [-s <since>] [-c <cursor>]
```

### Parameters

- `-user`: User address (required)
- `-token`: User authentication token (required)
- `-a`: Device address to fetch events from (optional if domain is provided)
- `-d`: Domain to fetch events from (optional if device address is provided)
- `-s`: Start fetching events from this date/time (e.g., '2024-01-02 15:04:05')
- `-c`: Pagination cursor to resume from a previous fetch

### Example

```bash
# Fetch all bstates events from a specific device
go run . -user myuser@domain.com -token mytoken -a device123

# Fetch events from all devices in a domain since a specific date
go run . -user myuser@domain.com -token mytoken -d mydomain -s "2024-01-02 15:04:05"
```

## Event Structure

Each event contains:
- `UID`: Unique event identifier
- `Type`: Mimetype (contains the schema ID)
- `SourceId`: Identifies the event source within the device
- `Meta`: Optional metadata map
- `Payload`: Binary-encoded state batch
- `Timestamp`: When the event was generated
- `Domain` and `Address`: Event origin

## Notes

- The example uses long polling with a 1-minute timeout
- Schemas are cached locally to avoid repeated fetches
- Each state in the batch is converted to JSON for display