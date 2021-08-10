# leif

`leif` is a service designed to take a set of GitHub repositories to 
find and track their configuration files for service level objective (SLO) rules,
periodically syncing to update the SLO rules, which are exposed over an API.

![leif](https://vignette.wikia.nocookie.net/animalcrossing/images/1/1c/Leif_NH.png/revision/latest/top-crop/width/360/height/360?cb=20200630055201)


### SLO rules configuration

`leif` expects the configuration of the SLO rules to be defined in a JSON file named `issue_slo_rules.json`

`leif` looks for the config file for the repository in the following places:
1. In the repository itself: `.github/issue_slo_rules.json`
2. If there is no config file in the respository, `leif` looks for the config file at the owner level: `<ownername>/.github/issue_slo_rules.json`

Note: an empty config file opts-out of SLO tracking
