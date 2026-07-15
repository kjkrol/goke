// Package bulk defines the contract for chunk-level bulk operations on
// entities: the data that describes a batch ([ChunkSnapshot] plus the
// accompanying entity IDs) and the interfaces that execute it ([Migrator]).
//
// The package holds no logic — it is the shared vocabulary spoken by the
// producers, transports, and executors of bulk commands.
package bulk
