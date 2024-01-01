package config

type DisableDeletionProtection struct {
	RDSInstance         bool `yaml:"RDSInstance"`
	EC2Instance         bool `yaml:"EC2Instance"`
	CloudformationStack bool `yaml:"CloudformationStack"`
	ELBv2               bool `yaml:"ELBv2"`
	QLDBLedger          bool `yaml:"QLDBLedger"`
}

type FeatureFlags struct {
	DisableDeletionProtection        DisableDeletionProtection `yaml:"disable-deletion-protection"`
	DisableEC2InstanceStopProtection bool                      `yaml:"disable-ec2-instance-stop-protection"`
	ForceDeleteLightsailAddOns       bool                      `yaml:"force-delete-lightsail-addons"`
}
