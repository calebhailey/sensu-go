package silenced

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// InfoCommand defines new silenced info command
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info [ID]",
		Short:        "show detailed silenced information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			id, err := getID(cmd, args)
			if err != nil {
				return err
			}
			r, err := cli.Client.FetchSilenced(id)
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
			return printToList(r, cmd.OutOrStdout())
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	cmd.Flags().StringP("subscription", "s", "*", "name of the silenced subscription")
	cmd.Flags().StringP("check", "c", "*", "name of the silenced check")

	return cmd

}

func expireTime(beginTS, expireSeconds int64) time.Duration {
	begin := time.Unix(beginTS, 0)
	expire := time.Duration(expireSeconds) * time.Second
	if time.Now().Before(begin) {
		return (expire - time.Until(begin)).Truncate(time.Second)
	}
	return time.Duration(expireSeconds) * time.Second
}

func printToList(r *types.Silenced, writer io.Writer) error {
	cfg := &list.Config{
		Title: r.ID,
		Rows: []*list.Row{
			{
				Label: "Expire",
				Value: expireTime(r.Begin, r.Expire).String(),
			},
			{
				Label: "ExpireOnResolve",
				Value: fmt.Sprintf("%t", r.ExpireOnResolve),
			},
			{
				Label: "Creator",
				Value: r.Creator,
			},
			{
				Label: "Check",
				Value: r.Check,
			},
			{
				Label: "Reason",
				Value: r.Reason,
			},
			{
				Label: "Subscription",
				Value: r.Subscription,
			},
			{
				Label: "Organization",
				Value: r.Organization,
			},
			{
				Label: "Environment",
				Value: r.Environment,
			},
		},
	}

	if time.Now().Before(time.Unix(r.Begin, 0)) {
		extraRows := []*list.Row{{
			Label: "Begin",
			Value: time.Unix(r.Begin, 0).Format(time.RFC822),
		}}
		cfg.Rows = append(extraRows, cfg.Rows...)
	}

	return list.Print(writer, cfg)
}
