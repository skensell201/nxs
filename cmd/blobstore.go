package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nexspence/nxs/internal/client"
)

var blobstoreCmd = &cobra.Command{
	Use:   "blobstore",
	Short: "Manage blob stores",
}

// humanBytes renders a byte count in the largest unit that keeps it above 1.
func humanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// blobStoreRows renders blob stores as table rows matching blobstoreListCmd's header.
func blobStoreRows(stores []client.BlobStore) [][]string {
	rows := make([][]string, 0, len(stores))
	for _, s := range stores {
		quota := "unlimited"
		if s.QuotaBytes != nil {
			quota = humanBytes(*s.QuotaBytes)
		}
		rows = append(rows, []string{s.Name, s.Type, humanBytes(s.UsedBytes), quota})
	}
	return rows
}

var blobstoreListCmd = &cobra.Command{
	Use:   "list",
	Short: "List blob stores",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := requireClient(); err != nil {
			return err
		}
		stores, err := nxsClient.BlobStoreList()
		if err != nil {
			return err
		}
		if flagJSON {
			printer.JSON(stores)
			return nil
		}
		printer.Table([]string{"NAME", "TYPE", "USED", "QUOTA"}, blobStoreRows(stores))
		return nil
	},
}

var blobstoreInfoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show blob store details",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		if err := requireClient(); err != nil {
			return err
		}
		bs, err := nxsClient.BlobStoreInfo(args[0])
		if err != nil {
			return err
		}
		if flagJSON {
			printer.JSON(bs)
			return nil
		}
		quota := "unlimited"
		if bs.QuotaBytes != nil {
			quota = humanBytes(*bs.QuotaBytes)
		}
		printer.Table([]string{"FIELD", "VALUE"}, [][]string{
			{"Name", bs.Name},
			{"Type", bs.Type},
			{"Used", humanBytes(bs.UsedBytes)},
			{"Quota", quota},
		})
		return nil
	},
}

var blobstoreCompactCmd = &cobra.Command{
	Use:   "compact <name>",
	Short: "Run garbage collection on a blob store (remove unreferenced blobs)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireClient(); err != nil {
			return err
		}
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		minAge, _ := cmd.Flags().GetString("min-age")
		res, err := nxsClient.BlobStoreCompact(args[0], dryRun, minAge)
		if err != nil {
			return err
		}
		if flagJSON {
			printer.JSON(res)
			return nil
		}
		verb := "Collected"
		if res.DryRun {
			verb = "Would collect"
		}
		printer.Success(fmt.Sprintf("%s %d orphan(s), %d bytes freed (%d blobs scanned) in %q",
			verb, res.Orphans, res.FreedBytes, res.ScannedBlobs, res.Store))
		return nil
	},
}

func init() {
	blobstoreCompactCmd.Flags().Bool("dry-run", false, "report orphans without deleting them")
	blobstoreCompactCmd.Flags().String("min-age", "", "only collect orphans older than this (e.g. 24h); overrides server default")
	blobstoreCmd.AddCommand(blobstoreListCmd)
	blobstoreCmd.AddCommand(blobstoreInfoCmd)
	blobstoreCmd.AddCommand(blobstoreCompactCmd)
	rootCmd.AddCommand(blobstoreCmd)
}
