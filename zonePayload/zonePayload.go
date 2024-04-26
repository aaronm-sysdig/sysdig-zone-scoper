package zonePayload

import (
	"fmt"
	"github.com/aaronm-sysdig/sysdig-zone-scoper/sysdighttp"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

// ZonePayload now maps directly to a map with zone names as keys
type ZonePayload struct {
	Zones map[string]Zone
}

type ZoneData struct {
	Data []Zone   `json:"data"`
	Page PageInfo `json:"page"`
}

type Zone struct {
	Author         string  `json:"author"`
	Description    string  `json:"description"`
	ID             int     `json:"id"`
	IsSystem       bool    `json:"isSystem"`
	LastModifiedBy string  `json:"lastModifiedBy"`
	LastUpdated    int64   `json:"lastUpdated"`
	Name           string  `json:"name"`
	Scopes         []Scope `json:"scopes"`
	Keep           bool
}

type CreateZone struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Scopes      []Scope `json:"scopes"`
}

type UpdateZone struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	Scopes []Scope `json:"scopes"`
}

type Scope struct {
	Rules      string `json:"rules"`
	TargetType string `json:"targetType"`
}

type PageInfo struct {
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Total    int     `json:"total"`
}

// NewZonePayload creates a new instance of ZonePayload with initialized map.
func NewZonePayload() *ZonePayload {
	return &ZonePayload{
		Zones: make(map[string]Zone),
	}
}

func (p *ZonePayload) GetZones(logger *logrus.Logger, configZones *sysdighttp.SysdigRequestConfig) (err error) {
	var objFetchZonesResponse *http.Response
	configZones.Path = "/platform/v1/zones"

	if objFetchZonesResponse, err = sysdighttp.SysdigRequest(logger, *configZones); err != nil {
		logger.Errorf("Could not retrieve zones: %v", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(objFetchZonesResponse.Body)

	var zoneData ZoneData
	if err = sysdighttp.ResponseBodyToJson(objFetchZonesResponse, &zoneData); err != nil {
		logger.Errorf("Could not unmarshal zones payload: %v", err)
		return err
	}

	p.Zones = make(map[string]Zone)
	for _, zone := range zoneData.Data {
		p.Zones[zone.Name] = zone
	}

	logger.Debugf("Successfully retrieved '%d' zones", len(p.Zones))
	return nil
}

// CreateNewZone sends a request to create a new zone.
func (p *ZonePayload) CreateNewZone(logger *logrus.Logger, configNewzone *sysdighttp.SysdigRequestConfig, createZone *CreateZone) (zone *Zone, err error) {
	configNewzone.Path = "/platform/v1/zones"
	configNewzone.Method = "POST"
	configNewzone.JSON = createZone

	response, err := sysdighttp.SysdigRequest(logger, *configNewzone)
	if err != nil {
		logger.Errorf("Failed to create zone: %v", err)
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	// Optionally decode the response if the API returns the created zone's details
	var createdZone Zone
	if err := sysdighttp.ResponseBodyToJson(response, &createdZone); err != nil {
		logger.Errorf("Could not decode response: %v", err)
		return nil, err
	}
	return &createdZone, nil
}

// UpdateZone sends a request to update an existing zone..
func (p *ZonePayload) UpdateZone(logger *logrus.Logger, configUpdateZone *sysdighttp.SysdigRequestConfig, updateZone *UpdateZone) error {
	configUpdateZone.Path = fmt.Sprintf("/platform/v1/zones/%d", updateZone.ID)
	configUpdateZone.Method = "PUT"
	configUpdateZone.JSON = updateZone

	response, err := sysdighttp.SysdigRequest(logger, *configUpdateZone)
	if err != nil {
		logger.Errorf("Failed to update zone: %v", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	// Decode the response if the API returns the zone's details
	var createdZone Zone
	if err := sysdighttp.ResponseBodyToJson(response, &createdZone); err != nil {
		logger.Errorf("Could not decode response: %v", err)
		return err
	}
	p.Zones[createdZone.Name] = createdZone
	return nil
}
