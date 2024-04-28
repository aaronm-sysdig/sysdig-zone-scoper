package teamPayload

import (
	"fmt"
	"github.com/aaronm-sysdig/sysdig-zone-scoper/sysdighttp"
	"github.com/sirupsen/logrus"
	"net/http"
)

type TeamBase struct {
	Page interface{}
	Data []CreateTeamPayload
}

type CreateTeamPayload struct {
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

func (tz *CreateTeamPayload) CreateTeamZoneMapping(logger *logrus.Logger,
	teamName string,
	zoneIds []int64,
	configCreateTeam *sysdighttp.SysdigRequestConfig,
	templateTeamPayload *CreateTeamPayload) (err error) {
	var objCreateTeamResponse *http.Response

	templateTeamPayload.ZoneIds = zoneIds
	templateTeamPayload.Name = teamName
	templateTeamPayload.Description = fmt.Sprintf("%s \nTeamName: '%s'", templateTeamPayload.Description, teamName)
	templateTeamPayload.Product = "secure"

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
