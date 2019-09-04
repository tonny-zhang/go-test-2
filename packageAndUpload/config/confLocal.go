package config

// ConfLocal conf for local
type ConfLocal struct {
	DirLocal    string
	DirLocalTmp string
	Host        string
	Port        int
	User        string
	Pwd         string
	DirRemote   string
	VersionGig  int
	ExtExclude  []string
	QnBucket    string
	QnKey       string
	QnSecret    string
	QnZone      string
}
