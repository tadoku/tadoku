package featureflags

// BooleanFlag identifies a boolean feature flag whose key and safe default are
// owned by this registry. Call sites cannot supply either value.
type BooleanFlag uint8

const (
	ReleaseLogEntryV2 BooleanFlag = iota
)

type booleanDefinition struct {
	key         string
	safeDefault bool
}

func (f BooleanFlag) definition() booleanDefinition {
	switch f {
	case ReleaseLogEntryV2:
		return booleanDefinition{
			key:         "release.log-entry-v2",
			safeDefault: false,
		}
	default:
		return booleanDefinition{}
	}
}

// Key returns the provider key owned by the registry.
func (f BooleanFlag) Key() string {
	return f.definition().key
}

// SafeDefault returns the behavior-preserving value used whenever a decision
// cannot be obtained safely.
func (f BooleanFlag) SafeDefault() bool {
	return f.definition().safeDefault
}
