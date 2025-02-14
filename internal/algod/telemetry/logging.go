package telemetry

// Config represents the configuration settings for telemetry, including enabling, logging, reporting, and user details.
// TODO: replace this with logging.TelemetryConfig
type Config struct {
	Enable             bool
	SendToLog          bool
	URI                string
	Name               string
	GUID               string
	FilePath           string
	UserName           string
	Password           string
	MinLogLevel        int
	ReportHistoryLevel int
}

// IsEqual compares two Config objects and returns true if all their fields have the same values, otherwise false.
func (c Config) IsEqual(conf Config) bool {
	return c.Enable == conf.Enable &&
		c.SendToLog == conf.SendToLog &&
		c.URI == conf.URI &&
		c.Name == conf.Name &&
		c.GUID == conf.GUID &&
		c.FilePath == conf.FilePath &&
		c.UserName == conf.UserName &&
		c.Password == conf.Password &&
		c.MinLogLevel == conf.MinLogLevel &&
		c.ReportHistoryLevel == conf.ReportHistoryLevel
}

// MergeLogConfigs merges two Config objects, with non-zero and non-default fields in 'b' overriding those in 'a'.
func MergeLogConfigs(a Config, b Config) Config {
	merged := a

	if b.Enable != a.Enable {
		merged.Enable = b.Enable
	}
	if b.SendToLog != a.SendToLog {
		merged.SendToLog = b.SendToLog
	}
	if b.URI != "" && b.URI != a.URI {
		merged.URI = b.URI
	}
	if b.Name != "" && b.Name != a.Name {
		merged.Name = b.Name
	}
	if b.GUID != "" && b.GUID != a.GUID {
		merged.GUID = b.GUID
	}
	if b.FilePath != "" && b.FilePath != a.FilePath {
		merged.FilePath = b.FilePath
	}
	if b.UserName != "" && b.UserName != a.UserName {
		merged.UserName = b.UserName
	}
	if b.Password != "" && b.Password != a.Password {
		merged.Password = b.Password
	}
	if b.MinLogLevel != 0 && b.MinLogLevel != a.MinLogLevel {
		merged.MinLogLevel = b.MinLogLevel
	}
	if b.ReportHistoryLevel != 0 && b.ReportHistoryLevel != a.ReportHistoryLevel {
		merged.ReportHistoryLevel = b.ReportHistoryLevel
	}

	return merged
}
