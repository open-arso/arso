package output

import (
	"fmt"

	"github.com/openarso/arso/apps/internal/node"
	"github.com/spf13/cobra"
)

func printNodeStatusText(cmd *cobra.Command, status node.Status) {
	fmt.Fprintf(cmd.OutOrStdout(), "Go routines:  %d\n", status.Runtime.GoRoutines)
	fmt.Fprintf(cmd.OutOrStdout(), "Go version:   %s\n", status.Runtime.GoVersion)
	fmt.Fprintf(cmd.OutOrStdout(), "OS:           %s\n", status.Runtime.OS)
	fmt.Fprintf(cmd.OutOrStdout(), "Arch:         %s\n", status.Runtime.Arch)
	fmt.Fprintf(cmd.OutOrStdout(), "State:        %s\n", status.State)
	fmt.Fprintf(cmd.OutOrStdout(), "Started At:   %s\n", status.StartedAt)
	fmt.Fprintf(cmd.OutOrStdout(), "Uptime:       %s\n", status.Uptime)
	fmt.Fprintf(cmd.OutOrStdout(), "Memory (MB):  %d / %d (%.2f %%)\n", status.Memory.UsedMB, status.Memory.TotalMB, status.Memory.Percent)
	fmt.Fprintf(cmd.OutOrStdout(), "Disk (GB):    %d / %d (%.2f %%)\n", status.Disk.UsedGB, status.Disk.TotalGB, status.Disk.Percent)
	fmt.Fprintf(cmd.OutOrStdout(), "CPU:          %.2f\n", status.CPU.UsagePercent)
}

func PrintNodeStatus(cmd *cobra.Command, result node.Status, output string) error {
	switch output {
	case Text:
		printNodeStatusText(cmd, result)
		return nil

	case JSON:
		return PrintJSON(cmd, result)

	case NDJSON:
		return PrintNDJSON(cmd, []node.Status{result})

	default:
		return fmt.Errorf("unhandled output format %q", output)
	}
}
