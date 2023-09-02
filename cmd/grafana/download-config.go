package grafana

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vegaprotocol/vega-monitoring/clients/grafana"
)

type DownloadConfigArgs struct {
	*GrafanaArgs
	NoAlert       bool
	NoDashboards  bool
	NoDataSources bool
	StorageDir    string
}

var downloadConfigArgs DownloadConfigArgs

// downloadConfigCmd represents the downloadConfig command
var downloadConfigCmd = &cobra.Command{
	Use:   "download-config",
	Short: "Download Grafana config",
	Long:  `Download Grafana config`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RunDownloadConfig(downloadConfigArgs); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	GrafanaCmd.AddCommand(downloadConfigCmd)
	downloadConfigArgs.GrafanaArgs = &grafanaArgs

	downloadConfigCmd.PersistentFlags().BoolVar(&downloadConfigArgs.NoDashboards, "no-dashboards", false, "Skip downloading dashboards config")
	downloadConfigCmd.PersistentFlags().BoolVar(&downloadConfigArgs.NoDataSources, "no-data-source", false, "Skip downloading Data Source config")
	downloadConfigCmd.PersistentFlags().BoolVar(&downloadConfigArgs.NoAlert, "no-alert", false, "Skip downloading Alert config")
	downloadConfigCmd.PersistentFlags().StringVar(&downloadConfigArgs.StorageDir, "config-dir", "grafana", "Directory where to put all data")
}

func RunDownloadConfig(args DownloadConfigArgs) error {

	grafanaClient := grafana.NewGrafanaClient(args.ApiURL, args.ApiToken)

	if !args.NoDashboards {
		if err := downloadDashboardsConfig(grafanaClient, filepath.Join(args.StorageDir, "dashboards")); err != nil {
			return err
		}
	}

	if !args.NoAlert {
		if err := downloadAlertConfig(grafanaClient, filepath.Join(args.StorageDir, "alerts")); err != nil {
			return err
		}
	}

	if !args.NoDataSources {
		if err := downloadDataSourcesConfig(grafanaClient, args.StorageDir); err != nil {
			return err
		}
	}

	return nil
}

func downloadDashboardsConfig(
	grafanaCleint *grafana.GrafanaClient,
	dashboardsConfigDir string,
) error {
	fmt.Printf("Downloading 'Dashboards' config\n")
	dashboardsMeta, err := grafanaCleint.GetDashboardList()
	if err != nil {
		return err
	}
	dashboardJSONs := map[string][]byte{}
	for _, meta := range dashboardsMeta {
		filename := fmt.Sprintf("%s - %s.json", meta.FolderTitle, meta.Title)
		if len(meta.FolderTitle) == 0 {
			filename = fmt.Sprintf("General - %s.json", meta.Title)
		}
		dashboardJSONs[filename], err = grafanaCleint.GetDashboardAsPrettyJSON(meta.UID)
		if err != nil {
			return fmt.Errorf("failed to get dashboard for %s, %w", meta.Title, err)
		}
	}

	fmt.Printf("Storing 'Dashboards' cofnig in %s\n", dashboardsConfigDir)
	if err := os.RemoveAll(dashboardsConfigDir); err != nil {
		return err
	}
	if err := os.MkdirAll(dashboardsConfigDir, 0755); err != nil {
		return err
	}

	for filename, dashboardJSON := range dashboardJSONs {
		if err = os.WriteFile(filepath.Join(dashboardsConfigDir, filename), dashboardJSON, 0644); err != nil {
			return err
		}
	}

	fmt.Printf("Successfully updated 'Dashboards' config\n")
	return nil
}

func downloadAlertConfig(
	grafanaCleint *grafana.GrafanaClient,
	alertConfigDir string,
) error {
	fmt.Printf("Downloading 'Alert' config\n")
	rulesJSON, err := grafanaCleint.GetAlertRulesAsPrettyJSON()
	if err != nil {
		return err
	}
	rulesYAML, err := grafanaCleint.GetAlertRulesAsYAML()
	if err != nil {
		return err
	}
	contactPointsJSON, err := grafanaCleint.GetAlertContactPointsAsPrettyJSON()
	if err != nil {
		return err
	}
	// contactPointsYAML, err := grafanaCleint.GetAlertContactPointsAsYAML()
	// if err != nil {
	// 	return err
	// }
	notificationPoliciesJSON, err := grafanaCleint.GetAlertNotificationPoliciesAsPrettyJSON()
	if err != nil {
		return err
	}
	// notificationPoliciesYAML, err := grafanaCleint.GetAlertNotificationPoliciesAsYAML()
	// if err != nil {
	// 	return err
	// }
	muteTimingsJSON, err := grafanaCleint.GetAlertMuteTimingsAsPrettyJSON()
	if err != nil {
		return err
	}
	templatesJSON, err := grafanaCleint.GetAlertTemplatesAsPrettyJSON()
	if err != nil {
		return err
	}

	fmt.Printf("Storing 'Alert' cofnig in %s\n", alertConfigDir)
	if err := os.RemoveAll(alertConfigDir); err != nil {
		return err
	}
	if err := os.MkdirAll(alertConfigDir, 0755); err != nil {
		return err
	}

	if err = os.WriteFile(filepath.Join(alertConfigDir, "alert-rules.json"), rulesJSON, 0644); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(alertConfigDir, "alert-rules.yaml"), rulesYAML, 0644); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(alertConfigDir, "contact-points.json"), contactPointsJSON, 0644); err != nil {
		return err
	}
	// if err = os.WriteFile(filepath.Join(alertConfigDir, "contact-points.yaml"), contactPointsYAML, 0644); err != nil {
	// 	return err
	// }
	if err = os.WriteFile(filepath.Join(alertConfigDir, "notification-policies.json"), notificationPoliciesJSON, 0644); err != nil {
		return err
	}
	// if err = os.WriteFile(filepath.Join(alertConfigDir, "notification-policies.yaml"), notificationPoliciesYAML, 0644); err != nil {
	// 	return err
	// }
	if err = os.WriteFile(filepath.Join(alertConfigDir, "mute-timings.json"), muteTimingsJSON, 0644); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(alertConfigDir, "alert-templates.json"), templatesJSON, 0644); err != nil {
		return err
	}

	fmt.Printf("Successfully updated 'Alert' config\n")

	return nil
}

func downloadDataSourcesConfig(
	grafanaCleint *grafana.GrafanaClient,
	configHomeDir string,
) error {
	fmt.Printf("Downloading 'Data Sources' config\n")
	dataSourcesJSON, err := grafanaCleint.GetDataSourcesAsPrettyJSON()
	if err != nil {
		return err
	}

	filename := filepath.Join(configHomeDir, "data-sources.json")
	fmt.Printf("Storing 'Data Sources' cofnig in %s\n", filename)
	if err = os.WriteFile(filename, dataSourcesJSON, 0644); err != nil {
		return err
	}

	fmt.Printf("Successfully updated 'Data Sources' config\n")
	return nil
}
