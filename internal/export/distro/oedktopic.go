/*
* OEDK topics
*/

package distro

// Publish
const (
	INIT_TOPIC			=	"agentruntime/controlling/command/init"
	LOGLEVEL_TOPIC		=	"agentruntime/controlling/command/loglevel"
	STOP_TOPIC			=	"agentruntime/controlling/command/stop"
	CONFIGUPDATE_TOPIC	=	"boxmanager/monitoring/opresult/configupdate"
	OFFBOARD_TOPIC		=	"boxmanager/monitoring/opresult/offboard"
	VERSION_TOPIC		=	"boxmanager/monitoring/softwareinformation/version"
	UPLOADDATA_TOPIC	=	"runtime/inject/data/timeseries/{protocol}/{dataSourceId}"
	UPLOADDIAG_TOPIC	=	"runtime/inject/diag/timeseries/{protocol}/{dataSourceId}"
)

// Subscribe
const (
	CONNECTION_TOPIC		=	"agentruntime/monitoring/diagnostic/connection"
	ONBOARDING_TOPIC		=	"agentruntime/monitoring/diagnostic/onboarding"
	BUFFER_TOPIC			=	"agentruntime/monitoring/diagnostic/buffer"
	DATA_TOPIC				=	"agentruntime/monitoring/diagnostic/data"
	CLOCKSKEW_TOPIC			=	"agentruntime/monitoring/clockskew"
	INITINFO_TOPIC			=	"agentruntime/monitoring/opresult/init"
	STOPINFO_TOPIC			=	"agentruntime/monitoring/opresult/stop"
	CONFIGINFO_TOPIC		=	"cloud/monitoring/update/configuration"
	CONFIGPROINFO_TOPIC		=	"cloud/monitoring/update/configuration/{protocol}"
	CONFIGAPPINFO_TOPIC		=	"cloud/monitoring/update/configuration/{app_id}"
	STOPUPLOAD_TOPIC		=	"runtime/data/timeseries/stop/{protocol}/{dataSourceId}"
)