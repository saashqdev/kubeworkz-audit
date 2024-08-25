# v1.2.0
2022-4-26
##Enhance:
- replace kubeworkz version v1.1.0 to v1.2.0
- replace jwt version v3.2.0 to v3.2.1
## Bugfix:
- Fix the paging problem of querying audit logs
## Dependency:
- kubeworkz 1.2.0

# v1.1.0
2021-12-16
##Feature: 
- add apis which receive generic and webconsole audit logs

# v1.0.1
2021-9-3
## Bugfix:
- Modify the switch logic: When the audit function in hotplug is on and the user configures or built-in ES, the audit function is on, otherwise it is off

# v1.0.0
## Features:
- Receive audit information from Kubeworkz and K8s and send it to ES
- Query and download audit information