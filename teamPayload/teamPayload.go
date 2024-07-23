package teamPayload

import (
	"fmt"
	"github.com/aaronm-sysdig/sysdig-zone-scoper/config"
	"github.com/aaronm-sysdig/sysdig-zone-scoper/sysdighttp"
	"github.com/sirupsen/logrus"
	"net/http"
)

type TeamBase struct {
	Page interface{}
	Data []TeamPayload
}

type TeamPayload struct {
	AdditionalTeamPermissions AdditionalPermissions `json:"additionalTeamPermissions"`
	CustomTeamRoleID          *int64                `json:"customTeamRoleId,omitempty"`
	Description               string                `json:"description"`
	IsAllZones                bool                  `json:"isAllZones"`
	IsDefaultTeam             bool                  `json:"isDefaultTeam"`
	Name                      string                `json:"name"`
	Scopes                    []Scope               `json:"scopes"`
	StandardTeamRole          string                `json:"standardTeamRole"`
	UiSettings                UISettings            `json:"uiSettings"`
	ZoneIds                   []int64               `json:"zoneIds"`
	Product                   string                `json:"product"`
	ID                        int64                 `json:"id',omitempty"`
	Version                   int64                 `json:"version,omitempty"`
}

type AdditionalPermissions struct {
	HasAgentCli             bool `json:"hasAgentCli"`
	HasAwsData              bool `json:"hasAwsData"`
	HasBeaconMetrics        bool `json:"hasBeaconMetrics"`
	HasInfrastructureEvents bool `json:"hasInfrastructureEvents"`
	HasRapidResponse        bool `json:"hasRapidResponse"`
	HasSysdigCaptures       bool `json:"hasSysdigCaptures"`
}

type Scope struct {
	Expression string `json:"expression"`
	Type       string `json:"type"`
}

type UISettings struct {
	Theme string `json:"theme"`
}

func (tb *TeamBase) GetTeamByName(logger *logrus.Logger,
	configGetTeamByName *sysdighttp.SysdigRequestConfig,
	teamName string) (err error) {

	var objGetTeamByNameResponse *http.Response
	configGetTeamByName.Path = fmt.Sprintf("/platform/v1/teams?filter=name:%s", teamName)

	if objGetTeamByNameResponse, err = sysdighttp.SysdigRequest(logger, *configGetTeamByName); err != nil {
		return err
	}

	if err = sysdighttp.ResponseBodyToJson(objGetTeamByNameResponse, tb); err != nil {
		logger.Errorf("Could not get team '%s'", teamName)
		return err
	}
	logger.Debugf("Returning %+v", *tb)
	return nil
}

func (tz *TeamPayload) CreateTeamZoneMapping(logger *logrus.Logger,
	appConfig *config.Configuration,
	teamName string,
	zoneIds []int64,
	configCreateTeam *sysdighttp.SysdigRequestConfig,
	templateTeamPayload *TeamPayload) (err error) {
	var objCreateTeamResponse *http.Response

	templateTeamPayload.ZoneIds = zoneIds
	templateTeamPayload.Name = teamName
	templateTeamPayload.Description = teamName

	configCreateTeam.Path = "/platform/v1/teams"
	configCreateTeam.JSON = templateTeamPayload
	configCreateTeam.Method = "POST"
	configCreateTeam.Headers = map[string]string{
		"Content-Type": "application/json",
	}

	if objCreateTeamResponse, err = sysdighttp.SysdigRequest(logger, *configCreateTeam); err != nil {
		return err
	}

	if err = sysdighttp.ResponseBodyToJson(objCreateTeamResponse, &tz); err != nil {
		logger.Error("Could not unmarshal new team payload")
		return err
	}
	return nil
}

func (tz *TeamPayload) UpdateTeamZoneMapping(logger *logrus.Logger,
	teamName string,
	zoneIds []int64,
	configUpdateTeam *sysdighttp.SysdigRequestConfig,
	templateTeamPayload *TeamPayload) (err error) {
	var objCreateTeamResponse *http.Response

	templateTeamPayload.ZoneIds = zoneIds
	templateTeamPayload.Name = teamName

	configUpdateTeam.Path = fmt.Sprintf("/platform/v1/teams/%d", templateTeamPayload.ID)
	configUpdateTeam.JSON = templateTeamPayload
	configUpdateTeam.Method = "PUT"
	configUpdateTeam.Headers = map[string]string{
		"Content-Type": "application/json",
	}

	if objCreateTeamResponse, err = sysdighttp.SysdigRequest(logger, *configUpdateTeam); err != nil {
		return err
	}

	if err = sysdighttp.ResponseBodyToJson(objCreateTeamResponse, &tz); err != nil {
		logger.Error("Could not unmarshal update team payload")
		return err
	}
	return nil
}
