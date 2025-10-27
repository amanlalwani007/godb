# GoDB

A simple append-only log-backed key-value database implementation in Go, following the principles from ["Build Your Own Database"](https://build-your-own.org/database/).

## Overview

GoDB is a lightweight, educational key-value store that demonstrates fundamental database concepts including:
- Append-only log structure for durability
- In-memory indexing for fast lookups
- Simple CLI interface for interaction
- File-based persistence

## Features

- **Append-Only Log**: All writes are appended to a log file, ensuring data durability
- **Key-Value Store**: Simple interface for storing and retrieving data
- **CLI Interface**: Interactive command-line interface for database operations
- **Persistence**: Data survives process restarts through file-based storage

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd godb

# Build the project
go build -o godb

# Run the database
./godb
```

## Usage

### Starting the Database

```bash
./godb
```

This will start the GoDB CLI with an interactive prompt.

### Available Commands

#### Set a Key-Value Pair
```
> set <key> <value>
```
Stores a value associated with the given key.

**Example:**
```
> set username alice
> set age 25
```

#### Get a Value
```
> get <key>
```
Retrieves the value associated with the given key.

**Example:**
```
> get username
alice
```

#### Delete a Key
```
> del <key>
```
Removes the key and its associated value from the database.

**Example:**
```
> del username
```

#### Compact the Log
```
> compact
```
Compacts the append-only log by removing deleted entries and consolidating the data file. This reduces disk space usage.

#### Exit
```
> exit
```
Closes the database connection and exits the CLI.

## Architecture

### Append-Only Log

GoDB uses an append-only log structure where:
1. All write operations (set/delete) are appended to a log file
2. The log is never modified in-place, only appended to
3. Read operations scan the log to find the latest value for a key
4. The entire log is read on startup to build the current state

### Data Format

The log file stores entries in a simple binary format:
- Each entry contains: operation type, key, and value
- Deleted keys are marked with a special tombstone entry
- The file grows over time until compaction is performed

### Compaction

The `compact` command:
- Reads the entire log file
- Filters out deleted entries and superseded values
- Writes a new compacted log file
- Replaces the old log with the compacted version

## Implementation Details

### Core Components

1. **Log File Manager**: Handles reading and writing to the append-only log
2. **CLI**: Command-line interface for user interaction
3. **Compaction Engine**: Manages log file compaction and optimization

### Performance Characteristics

- **Write Operations**: O(1) - appends to log file
- **Read Operations**: O(n) - scans log file from end to find latest value
- **Delete Operations**: O(1) - appends tombstone to log
- **Compaction**: O(n) - reads and rewrites entire log

## Limitations

This is an educational implementation with several limitations:

- **No Indexing**: Reads scan the entire log file (slow for large datasets)
- **No Concurrency Control**: Not safe for concurrent access
- **Single File**: All data in one log file
- **No Transactions**: Operations are not atomic across multiple keys
- **No Replication**: Single-node only, no high availability
- **Basic Compaction**: Stops the world during compaction

## Future Enhancements

Potential improvements for learning:

- [ ] Add in-memory index for fast lookups
- [ ] Add write-ahead logging (WAL) for crash recovery
- [ ] Implement B-tree indexing for range queries
- [ ] Add support for multiple data files (SSTables)
- [ ] Implement memtable and level compaction
- [ ] Add concurrent access with proper locking
- [ ] Support transactions with MVCC
- [ ] Add replication and distributed consensus

## Learning Resources

This project is based on the book:
- [Build Your Own Database](https://build-your-own.org/database/)

For understanding the concepts:
- Append-only logs and log-structured storage
- Indexing techniques
- Compaction strategies
- Database persistence and durability

## Project Structure

```
godb/
├── kv/
│   ├── kv.go             # Key-value operations
│   └── log.go            # Append-only log implementation
├── main.go               # Entry point and CLI
├── db.log                # Data file (created at runtime)
├── go.mod                # Go module file
└── README.md             # This file
```

## Contributing

This is an educational project. Feel free to:
- Experiment with the code
- Add new features
- Optimize performance
- Add tests

## License

[Specify your license here]

## Acknowledgments

- Based on the "Build Your Own Database" tutorial
- Inspired by real-world systems like Bitcask, LevelDB, and RocksDB