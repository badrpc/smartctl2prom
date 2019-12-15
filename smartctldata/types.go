// Package smartctldata provides data structures matching the output of
// smartctl utility and parsers suitable to read JSON and legacy text output
// of smartctl (text format is only partially supported).
package smartctldata

type Output struct {
	JsonFormatVersion  [2]int             `json:"json_format_version"`
	SmartCtl           Invocation         `json:"smartctl"`
	Device             Device             `json:"device"`
	ModelFamily        string             `json:"model_family"`
	ModelName          string             `json:"model_name"`
	SerialNumber       string             `json:"serial_number"`
	WWN                WWN                `json:"wwn"`
	FirmwareVersion    string             `json:"firmware_version"`
	UserCapacity       Capacity           `json:"user_capacity"`
	LogicalBlockSize   int64              `json:"logical_block_size"`
	PhysicalBlockSize  int64              `json:"physical_block_size"`
	InSmartCtlDatabase bool               `json:"in_smartctl_database"`
	ATAVersion         ATAVersion         `json:"ata_version"`
	SATAVersion        SATAVersion        `json:"sata_version"`
	InterfaceSpeed     InterfaceSpeed     `json:"interface_speed"`
	LocalTime          Time               `json:"local_time"`
	SMARTStatus        SMARTStatus        `json:"smart_status"`
	ATASMARTData       ATASMARTData       `json:"ata_smart_data"`
	ATASctCapabilities ATASctCapabilities `json:"ata_sct_capabilities"`
	ATASMARTAttributes ATASMARTAttributes `json:"ata_smart_attributes"`

	PowerOnTime     PowerOnTime `json:"power_on_time"`
	PowerCycleCount int64       `json:"power_cycle_count"`
	Temperature     Temperature `json:"temperature"`
}

type Invocation struct {
	Version      [2]int
	SvnRevision  string `json:"svn_revision"`
	PlatformInfo string `json:"platform_info"`
	BuildInfo    string
	Argv         []string
	ExitStatus   int
}

type Device struct {
	Name     string
	InfoName string `json:"info_name"`
	Type     string
	Protocol string
}

type WWN struct {
	NAA int64 `json:"naa"`
	OUI int64 `json:"oui"`
	ID  int64 `json:"id"`
}

type Capacity struct {
	Blocks int64
	Bytes  int64
}

type ATAVersion struct {
	Text  string `json:"string"`
	Major int64  `json:"major_value"`
	Minor int64  `json:"minor_value"`
}

type SATAVersion struct {
	Text  string `json:"string"`
	Value int64
}

type SpeedSpec struct {
	SATAValue      int64  `json:"sata_value"`
	Text           string `json:"string"`
	UnitsPerSecond int64  `json:"units_per_second"`
	BitsPerUnit    int64  `json:"bits_per_unit"`
}

type InterfaceSpeed struct {
	Max     SpeedSpec `json:"max"`
	Current SpeedSpec `json:"current"`
}

type Time struct {
	TimeT   int64  `json:"time_t"`
	AscTime string `json:"asctime"`
}

type SMARTStatus struct {
	Passed bool `json:"passed"`
}

type ATASMARTData struct {
	OfflineDataCollection OfflineDataCollection `json:"offline_data_collection"`
	SelfTest              SelfTestData          `json:"self_test"`
	Capabilities          SMARTCapabilities     `json:"capabilities"`
}

type OfflineDataCollection struct {
	Status            DataCollectionStatus `json:"status"`
	CompletionSeconds int64                `json:"completion_seconds"`
}

type DataCollectionStatus struct {
	Value int64  `json:"value"`
	Text  string `json:"string"`
}

type SelfTestData struct {
	Status         SelfTestStatus `json:"status"`
	PollingMinutes SelfTestTime   `json:"polling_minutes"`
}

type SelfTestStatus struct {
	Value  int64  `json:"value"`
	Text   string `json:"string"`
	Passed bool   `json:"passed"`
}

type SelfTestTime struct {
	Short      int64 `json:"short"`
	Extended   int64 `json:"extended"`
	Conveyance int64 `json:"conveyance"`
}

type SMARTCapabilities struct {
	Values                        []int64 `json:"values"`
	ExecOfflineImmediateSupported bool    `json:"exec_offline_immediate_supported"`
	OfflineIsAbortedUponNewCmd    bool    `json:"offline_is_aborted_upon_new_cmd"`
	OfflineSurfaceScanSupported   bool    `json:"offline_surface_scan_supported"`
	SelfTestsSupported            bool    `json:"self_tests_supported"`
	ConveyanceSelfTestSupported   bool    `json:"conveyance_self_test_supported"`
	SelectiveSelfTestSupported    bool    `json:"selective_self_test_supported"`
	AttributeAutosaveEnabled      bool    `json:"attribute_autosave_enabled"`
	ErrorLoggingSupported         bool    `json:"error_logging_supported"`
	GPLoggingSupported            bool    `json:"gp_logging_supported"`
}

type ATASctCapabilities struct {
	Value                         int64 `json:"value"`
	ErrorRecoveryControlSupported bool  `json:"error_recovery_control_supported"`
	FeatureControlSupported       bool  `json:"feature_control_supported"`
	DataTableSupported            bool  `json:"data_table_supported"`
}

type ATASMARTAttributes struct {
	Revision int32             `json:"revision"`
	Table    []*SMARTAttribute `json:"table"`
}

type SMARTAttribute struct {
	ID         int32                  `json:"id"`
	Name       string                 `json:"name"`
	Value      int32                  `json:"value"`
	Worst      int32                  `json:"worst"`
	Threshold  int32                  `json:"thresh"`
	WhenFailed string                 `json:"when_failed"`
	Flags      SMARTAttributeFlags    `json:"flags"`
	Raw        SMARTAttributeRawValue `json:"raw"`
}

type SMARTAttributeFlags struct {
	Value         int64  `json:"value"`
	Text          string `json:"string"`
	Prefailure    bool   `json:"prefailure"`
	UpdatedOnline bool   `json:"updated_online"`
	Performance   bool   `json:"performance"`
	ErrorRate     bool   `json:"error_rate"`
	EventCount    bool   `json:"event_count"`
	AutoKeep      bool   `json:"auto_keep"`
}

type SMARTAttributeRawValue struct {
	Value int64  `json:"value"`
	Text  string `json:"string"`
}

type PowerOnTime struct {
	Hours int64 `json:"hours"`
}

type Temperature struct {
	Current int64 `json:"current"`
}
