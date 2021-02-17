# TurtleDexDir
The TurtleDexDir module is responsible for creating and maintaining the directory
metadata information stored in the `.ttdxdir` files on disk. This includes all
disk interaction and metadata definition. These ttdxdirs represent directories on
the TurtleDex network.

## Structure of the TurtleDexDir
The TurtleDexDir is a dir on the TurtleDex network and the ttdxdir metadata is a JSON
formatted metadata file that contains aggregate and non-aggregate fields. The
aggregate fields are the totals of the ttdxdir and any sub ttdxdirs, or are
calculated based on all the values in the subtree. The non-aggregate fields are
information specific to the ttdxdir that is not an aggregate of the entire sub
directory tree

## Subsystems
The following subsystems help the TurtleDexDir module execute its responsibilities:
 - [Persistence Subsystem](#persistence-subsystem)
 - [File Format Subsystem](#file-format-subsystem)
 - [TurtleDexDirSet Subsystem](#ttdxdirset-subsystem)
 - [DirReader Subsystem](#dirreader-subsystem)

 ### Persistence Subsystem
 **Key Files**
- [persist.go](./persist.go)
- [persistwal.go](./persistwal.go)

The Persistence subsystem is responsible for the disk interaction with the
`.ttdxdir` files and ensuring safe and performant ACID operations by using the
[writeaheadlog](https://github.com/turtledex/writeaheadlog) package. There
are two WAL updates that are used, deletion and metadata updates.

The WAL deletion update removes all the contents of the directory including the
directory itself.

The WAL metadata update re-writes the entire metadata, which is stored as JSON.
This is used whenever the metadata changes and needs to be saved as well as when
a new ttdxdir is created.

**Exports**
 - `ApplyUpdates`
 - `CreateAndApplyTransaction`
 - `IsTurtleDexDirUpdate`
 - `New`
 - `LoadTurtleDexDir`
 - `UpdateMetadata`

**Inbound Complexities**
 - `callDelete` deletes a TurtleDexDir from disk
    - `TurtleDexDirSet.Delete` uses `callDelete`
 - `LoadTurtleDexDir` loads a TurtleDexDir from disk
    - `TurtleDexDirSet.open` uses `LoadTurtleDexDir`

### File Format Subsystem
 **Key Files**
- [ttdxdir.go](./ttdxdir.go)

The file format subsystem contains the type definitions for the TurtleDexDir
format and methods that return information about the TurtleDexDir.

**Exports**
 - `Deleted`
 - `Metatdata`
 - `TurtleDexPath`

### TurtleDexDirSet Subsystem
 **Key Files**
- [ttdxdirset.go](./ttdxdirset.go)

A TurtleDexDir object is threadsafe by itself, and to ensure that when a TurtleDexDir is
accessed by multiple threads that it is still threadsafe, TurtleDexDirs should always
be accessed through the TurtleDexDirSet. The TurtleDexDirSet was created as a pool of
TurtleDexDirs which is used by other packages to get access to TurtleDexDirEntries which are
wrappers for TurtleDexDirs containing some extra information about how many threads
are using it at a certain time. If a TurtleDexDir was already loaded the TurtleDexDirSet
will hand out the existing object, otherwise it will try to load it from disk.

**Exports**
 - `HealthPercentage`
 - `NewTurtleDexDirSet`
 - `Close`
 - `Delete`
 - `DirInfo`
 - `DirList`
 - `Exists`
 - `InitRootDir`
 - `NewTurtleDexDir`
 - `Open`
 - `Rename`

**Outbound Complexities**
 - `Delete` will use `callDelete` to delete the TurtleDexDir once it has been acquired
   in the set
 - `open` calls `LoadTurtleDexDir` to load the TurtleDexDir from disk

### DirReader Subsystem
**Key Files**
 - [dirreader.go](./dirreader.go)

The DirReader Subsystem creates the DirReader which is used as a helper to read
raw .ttdxdir from disk

**Exports**
 - `Close`
 - `Read`
 - `Stat`
 - `DirReader`
