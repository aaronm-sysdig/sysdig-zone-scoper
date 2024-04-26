# sysdig-zone-scoper
(interim readme until development is finished)

### Build

1) Clnne repo
2) Build / Run

### Configuration
Environment Variables

| Parameter           | Example                                |
|---------------------|----------------------------------------|
| GROUPING_LABEL      | `kubernetes.namespace.label.ZoneName`  |
| SECURE_TOKEN        | `ab211234-3ba6-4085-a579-9996272efa3b` |
| SYSDIG_API_ENDPOINT | `https://app.au1.sysdig.com`           |
| STATIC_ZONES        | zone to keep,my zone,another zone      |


```
GROUPING_LABEL=xxx> SECURE_API_TOKEN=xxx SYSDIG_API_ENDPOINT=xxx STATIC_ZONES="zone to keep,my zone, another zone" go run sysdig-zone-scoper.go
```

nb: You can also pass `--grouping-label` as a command line parameter if you wish