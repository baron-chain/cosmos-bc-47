package snapshot

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/baron-chain/cosmos-bc-47/server"
)

const (
	dumpCmdUse     = "dump <height> <format>"
	dumpCmdShort   = "Dump Baron Chain snapshot as portable archive"
	dumpCmdLong    = `Export a Baron Chain snapshot to a portable gzipped tar archive.
The archive will contain the snapshot metadata and all associated chunk files.`
	dumpCmdExample = `  # Dump snapshot at height 1000000 with format 1
  barond snapshots dump 1000000 1

  # Dump snapshot with custom output file
  barond snapshots dump 1000000 1 -o custom_backup.tar.gz`

	defaultFileMode = 0o644
	flagOutput      = "output"
	flagOutputShort = "o"
)

type snapshotDumper struct {
	store     server.SnapshotStore
	height    uint64
	format    uint32
	outputPath string
}

func DumpArchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     dumpCmdUse,
		Short:   dumpCmdShort,
		Long:    dumpCmdLong,
		Example: dumpCmdExample,
		Args:    cobra.ExactArgs(2),
		RunE:    runDumpCmd,
	}

	cmd.Flags().StringP(flagOutput, flagOutputShort, "", "Output file path")
	return cmd
}

func runDumpCmd(cmd *cobra.Command, args []string) error {
	height, err := parseHeight(args[0])
	if err != nil {
		return fmt.Errorf("invalid height: %w", err)
	}

	format, err := parseFormat(args[1])
	if err != nil {
		return fmt.Errorf("invalid format: %w", err)
	}

	ctx := server.GetServerContextFromCmd(cmd)
	store, err := server.GetSnapshotStore(ctx.Viper)
	if err != nil {
		return fmt.Errorf("failed to get snapshot store: %w", err)
	}

	outputPath, err := cmd.Flags().GetString(flagOutput)
	if err != nil {
		return err
	}
	if outputPath == "" {
		outputPath = fmt.Sprintf("%d-%d.tar.gz", height, format)
	}

	dumper := &snapshotDumper{
		store:      store,
		height:     height,
		format:     format,
		outputPath: outputPath,
	}

	if err := dumper.dump(); err != nil {
		return fmt.Errorf("failed to dump snapshot: %w", err)
	}

	cmd.Printf("Successfully dumped snapshot to %s\n", outputPath)
	return nil
}

func (d *snapshotDumper) dump() error {
	snapshot, err := d.store.Get(d.height, d.format)
	if err != nil {
		return fmt.Errorf("failed to get snapshot: %w", err)
	}
	if snapshot == nil {
		return fmt.Errorf("snapshot at height %d format %d doesn't exist", d.height, d.format)
	}

	snapshotBytes, err := snapshot.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	file, err := os.Create(d.outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	gzipWriter, err := gzip.NewWriterLevel(file, gzip.BestSpeed)
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %w", err)
	}
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	if err := d.writeSnapshotMetadata(tarWriter, snapshotBytes); err != nil {
		return err
	}

	if err := d.writeChunkFiles(tarWriter, snapshot.Chunks); err != nil {
		return err
	}

	return nil
}

func (d *snapshotDumper) writeSnapshotMetadata(tw *tar.Writer, data []byte) error {
	header := &tar.Header{
		Name: SnapshotFileName,
		Mode: defaultFileMode,
		Size: int64(len(data)),
	}

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write snapshot header: %w", err)
	}

	if _, err := tw.Write(data); err != nil {
		return fmt.Errorf("failed to write snapshot data: %w", err)
	}

	return nil
}

func (d *snapshotDumper) writeChunkFiles(tw *tar.Writer, chunks uint32) error {
	for i := uint32(0); i < chunks; i++ {
		chunkPath := d.store.PathChunk(d.height, d.format, i)
		
		if err := d.writeChunkFile(tw, chunkPath, i); err != nil {
			return fmt.Errorf("failed to write chunk %d: %w", i, err)
		}
	}
	return nil
}

func (d *snapshotDumper) writeChunkFile(tw *tar.Writer, path string, index uint32) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open chunk file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat chunk file: %w", err)
	}

	header := &tar.Header{
		Name: strconv.FormatUint(uint64(index), 10),
		Mode: defaultFileMode,
		Size: info.Size(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write chunk header: %w", err)
	}

	if _, err := io.Copy(tw, file); err != nil {
		return fmt.Errorf("failed to write chunk data: %w", err)
	}

	return nil
}
