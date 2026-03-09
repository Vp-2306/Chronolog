# ChronoLog

ChronoLog is a high-performance write-optimized key-value storage engine built in Go using the Log Structured Merge Tree (LSM Tree) architecture.

This project is being developed as part of a Database Systems project.

## Goals

ChronoLog is designed to efficiently handle high write workloads such as:

- system logs
- telemetry
- metrics
- event streams

## Architecture

ChronoLog follows the LSM Tree design:

Client
  │
  ▼
Storage Engine API
  │
  ├── Write Ahead Log (Durability)
  ├── MemTable (SkipList)
  ├── SSTables (Immutable Disk Storage)
  └── Compaction (Background Merging)

## Features (Planned)

- SkipList MemTable
- Write Ahead Log (WAL)
- Immutable SSTables
- Bloom Filters for faster reads
- Leveled Compaction

## Tech Stack

- Go (Golang)
- LSM Tree storage architecture
