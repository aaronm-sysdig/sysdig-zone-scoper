# sysdig-zone-scoper
(interim readme until development is finished)

### Build

1) Clnne repo
2) Build / Run

### Configuration
Environment Variables

| Parameter           | Description                                                   | Example                                |
|---------------------|---------------------------------------------------------------|----------------------------------------|
| GROUPING_LABEL      | Sets the label to group by                                    | `kubernetes.namespace.label.ZoneName`  |
| SECURE_TOKEN        | Sysdig secure API token                                       | `ab211234-3ba6-4085-a579-9996272efa3b` |
| SYSDIG_API_ENDPOINT | Sysdig API Endpoint                                           | `https://app.au1.sysdig.com`           |
| STATIC_ZONES        | Zones to keep and not delete even if we did not create them   | zone to keep,my zone,another zone      |
| TEAM_TEMPLATE_NAME  | Name of the team to use as a create template for teams        | TeamTemplate                           |
| TEAM_ZONE_MAPPING   | CSV file to use to map between 'Team' and 'Zones'             | mapping.csv                            |
| LOG_LEVEL           | Logging level for app                                         | Debug \|\| Info \|\| Error             |
| SILENT              | Run silently and do not prompt to confirm execution           | true                                   |
| MODE                | Determines execution mode. Values `team`, `zone` or `monitor` | monitor                                |
| TEAM_PREFIX         | Sets a team name prefix if required`                          |                                        |

** `CREATE_ZONES` and `CREATE_TEAMS` are mutually exclusive, don't pass both with true/false, just pass the one you want

### Commandline Paramter
`--silent/-s` Runs without the dry-run confirmation <br>
`--log-mode/-d` Sets the logging mode <br>
`--team-zone-mapping/-m` Sets the mapping CSV file to use <br>
`--grouping-label/-l` Sets the grouping label to use <br>
`--team-template-name/-e` Sets the team template name to use to use as a template for team creation (permissions etc) <br>
`--mode/-o` Sets execution mode
`--team-prefix/-t` Sets team name prefix (if any)
`--dryrun/-r` Runs in dry-run mode.  Will pretend to create but will not (enabled for Monitor mode only at the moment)

### `TEAM_ZONE_MAPPING` example
Once your zones are created, the next thing to do is create teams that use these zones.  the `TEAM_ZONE_MAPPING` configuration
achieves this. Pass it with either a `--team-zone-mapping` command line parameter or `TEAM_ZONE_MAPPING` environment variable
```
Team Name,Zone Label
Andrews Team,API Support, Webservers
Aarons Team,Development
```

### Exeecution example
```
CREATE_ZONES=true LOG_LEVEL=debug TEAM_ZONE_MAPPING=mapping.csv GROUPING_LABEL=xxx> SECURE_API_TOKEN=xxx SYSDIG_API_ENDPOINT=xxx STATIC_ZONES="zone to keep,my zone, another zone" go run sysdig-zone-scoper.go
```

### Exeecution example - Monitor

TEAM_PREFIX="TeamName: " TEAM_TEMPLATE_NAME=team-template-monitor MODE=monitor LOG_LEVEL=debug GROUPING_LABEL=kubernetes.namespae.label.support_group SECURE_API_TOKEN=xxx SYSDIG_API_ENDPOINT=xxx go run sysdig-zone-scoper.go
