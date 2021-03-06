package cron

var PackageConfiguration = config.NewConfiguration(
	&config.Section{
		ID:        "system",
		Label:     "",
		SortOrder: 0,
		Scope:     config.NewScopePerm(),
		Groups: config.GroupSlice{
			&config.Group{
				ID:        "cron",
				Label:     `Cron (Scheduled Tasks) - all the times are in minutes`,
				Comment:   `For correct URLs generated during cron runs please make sure that Web > Secure and Unsecure Base URLs are explicitly set.`,
				SortOrder: 15,
				Scope:     config.NewScopePerm(config.IDScopeDefault),
				Fields:    config.FieldSlice{},
			},
		},
	},
)
