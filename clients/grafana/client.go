package grafana

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type GrafanaClient struct {
	httpClient  *http.Client
	apiURL      string
	rateLimiter *rate.Limiter
	apiToken    string
}

func NewGrafanaClient(
	apiURL string,
	apiToken string,
) *GrafanaClient {
	return &GrafanaClient{
		apiURL:      apiURL,
		apiToken:    apiToken,
		rateLimiter: rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

//
// Alerts
//

type Alert struct {
	UID       string `json:"uid"`
	Title     string `json:"title"`
	RuleGroup string `json:"ruleGroup"`
}

func (c *GrafanaClient) GetAlertList() (alerts []Alert, err error) {
	err = c.GetJSON("/api/v1/provisioning/alert-rules", &alerts)
	return
}

func (c *GrafanaClient) GetAlertAsPrettyJSON(uid string) ([]byte, error) {
	return c.GetPrettyJSON(fmt.Sprintf("/api/v1/provisioning/alert-rules/%s", uid))
}

func (c *GrafanaClient) GetAllAlertRulesAsPrettyJSON() ([]byte, error) {
	return c.GetPrettyJSON("/api/v1/provisioning/alert-rules")
}
func (c *GrafanaClient) GetAlertRulesAsYAML() ([]byte, error) {
	return c.GetAny("/api/v1/provisioning/alert-rules/export")
}

func (c *GrafanaClient) GetAlertContactPointsAsPrettyJSON() ([]byte, error) {
	return c.GetPrettyJSON("/api/v1/provisioning/contact-points")
}
func (c *GrafanaClient) GetAlertContactPointsAsYAML() ([]byte, error) {
	return c.GetAny("/api/v1/provisioning/contact-points/export")
}

func (c *GrafanaClient) GetAlertNotificationPoliciesAsPrettyJSON() ([]byte, error) {
	return c.GetPrettyJSON("/api/v1/provisioning/policies")
}
func (c *GrafanaClient) GetAlertNotificationPoliciesAsYAML() ([]byte, error) {
	return c.GetAny("/api/v1/provisioning/policies/export")
}

func (c *GrafanaClient) GetAlertMuteTimingsAsPrettyJSON() ([]byte, error) {
	return c.GetPrettyJSON("/api/v1/provisioning/mute-timings")
}

func (c *GrafanaClient) GetAlertTemplatesAsPrettyJSON() ([]byte, error) {
	return c.GetPrettyJSON("/api/v1/provisioning/templates")
}

//
// Dashboards
//

type Dashboard struct {
	UID         string `json:"uid"`
	Title       string `json:"title"`
	FolderTitle string `json:"folderTitle"`
	Type        string `json:"type"`
}

func (c *GrafanaClient) GetDashboardList() ([]Dashboard, error) {
	var searchResult []Dashboard
	err := c.GetJSON("/api/search", &searchResult)
	if err != nil {
		return nil, err
	}
	var result []Dashboard
	for _, entry := range searchResult {
		if entry.Type == "dash-db" {
			result = append(result, entry)
		}
	}
	return result, nil
}

func (c *GrafanaClient) GetDashboardAsPrettyJSON(uid string) ([]byte, error) {
	return c.GetPrettyJSON(fmt.Sprintf("/api/dashboards/uid/%s", uid))
}

//
// Other
//

func (c *GrafanaClient) GetDataSourcesAsPrettyJSON() ([]byte, error) {
	return c.GetPrettyJSON("/api/datasources")
}
