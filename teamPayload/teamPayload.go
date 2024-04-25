package teamPayload

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
	Version                   int64                 `json:"version"`
	ZoneIds                   []string              `json:"zoneIds"`
	Id                        int64                 `json:"id"`
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
