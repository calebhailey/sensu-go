package check

import (
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/globals"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// InfoCommand defines new check info command
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info [ID]",
		Short:        "show detailed check information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch handlers from API
			checkID := args[0]
			r, err := cli.Client.FetchCheck(checkID)
			if err != nil {
				return err
			}

			// Determine the format to use to output the data
			var format string
			if format = helpers.GetChangedStringValueFlag("format", cmd.Flags()); format == "" {
				format = cli.Config.Format()
			}

			if format == "json" {
				return helpers.PrintJSON(r, cmd.OutOrStdout())
			}
			return printCheckToList(r, cmd.OutOrStdout())
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printCheckToList(r *types.CheckConfig, writer io.Writer) error {
	cfg := &list.Config{
		Title: r.Name,
		Rows: []*list.Row{
			{
				Label: "Name",
				Value: r.Name,
			},
			{
				Label: "Interval",
				Value: strconv.FormatInt(int64(r.Interval), 10),
			},
			{
				Label: "Command",
				Value: r.Command,
			},
			{
				Label: "Cron",
				Value: r.Cron,
			},
			{
				Label: "Timeout",
				Value: strconv.FormatInt(int64(r.Timeout), 10),
			},
			{
				Label: "TTL",
				Value: strconv.FormatInt(int64(r.Ttl), 10),
			},
			{
				Label: "Subscriptions",
				Value: strings.Join(r.Subscriptions, ", "),
			},
			{
				Label: "Handlers",
				Value: strings.Join(r.Handlers, ", "),
			},
			{
				Label: "Runtime Assets",
				Value: strings.Join(r.RuntimeAssets, ", "),
			},
			{
				Label: "Hooks",
				Value: globals.FormatHookLists(r.CheckHooks),
			},
			{
				Label: "Publish?",
				Value: strconv.FormatBool(r.Publish),
			},
			{
				Label: "Stdin?",
				Value: strconv.FormatBool(r.Stdin),
			},
			{
				Label: "Proxy Entity ID",
				Value: r.ProxyEntityID,
			},
			{
				Label: "Organization",
				Value: r.Organization,
			},
			{
				Label: "Environment",
				Value: r.Environment,
			},
			{
				Label: "Metric Format",
				Value: r.OutputMetricFormat,
			},
			{
				Label: "Metric Handlers",
				Value: strings.Join(r.OutputMetricHandlers, ", "),
			},
		},
	}

	return list.Print(writer, cfg)
}
