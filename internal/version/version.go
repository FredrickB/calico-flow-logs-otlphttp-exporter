package version

// Variables are set as part of the build using --ldflags "\
// -X github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/version.version=$(VERSION) \
// -X github.com/FredrickB/calico-flow-logs-otlphttp-exporter/v2/internal/version.goldmaneProtobufVersion=$(GOLDMANE_PROTO_VERSION) \
// "

const NO_VERSION_SET = "NO_VERSION_SET"

var (
	version                 string
	goldmaneProtobufVersion string
)

func Version() string {
	return returnNoVersionIfVersionIsNotSet(version)
}

func GoldmaneProtobufVersion() string {
	return returnNoVersionIfVersionIsNotSet(goldmaneProtobufVersion)
}

func returnNoVersionIfVersionIsNotSet(version string) string {
	if version == "" {
		return NO_VERSION_SET
	}
	return version
}
