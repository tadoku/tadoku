package featureflags

type BooleanFlag uint8

const (
	ReleaseLogEntryV2 BooleanFlag = iota
	IdentityV2
)

type booleanDefinition struct {
	key         string
	safeDefault bool
}

func (f BooleanFlag) definition() booleanDefinition {
	switch f {
	case ReleaseLogEntryV2:
		return booleanDefinition{
			key:         "release-log-entry-v2",
			safeDefault: false,
		}
	case IdentityV2:
		return booleanDefinition{key: "identity-v2", safeDefault: false}
	default:
		return booleanDefinition{}
	}
}

func (f BooleanFlag) Key() string {
	return f.definition().key
}

func (f BooleanFlag) SafeDefault() bool {
	return f.definition().safeDefault
}
