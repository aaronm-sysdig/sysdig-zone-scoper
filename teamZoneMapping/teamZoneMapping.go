package teamZoneMapping

import (
	"encoding/csv"
	"io"
)

// TeamZones maps a team name to a list of zone labels.
type TeamZones map[string][]string

// NewTeamZones initializes and returns a new TeamZones instance.
func NewTeamZones() *TeamZones {
	tz := make(TeamZones)
	return &tz
}

// ParseCSV parses CSV data from an io.Reader and fills the TeamZones map.
func (tz *TeamZones) ParseCSV(r io.Reader) error {
	csvReader := csv.NewReader(r)
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = -1

	// Skip the header row
	if _, err := csvReader.Read(); err != nil {
		return err // Return the error if unable to read the header (could be EOF or a different error)
	}

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break // End of file, stop reading
		}
		if err != nil {
			return err // Return any other error that might occur
		}

		// We assume there is at least one field per row (the team name)
		if len(record) < 1 {
			continue // Skip rows with no fields
		}

		teamName := record[0]    // The first column is the team name
		zoneLabels := record[1:] // All subsequent columns are zone labels

		// Appending the zones to the corresponding team
		(*tz)[teamName] = append((*tz)[teamName], zoneLabels...)
	}
	return nil
}
