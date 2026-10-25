package featureflags

type BooleanFlag uint8

const (
	ReleaseWebV3 BooleanFlag = iota
)

type booleanDefinition struct {
	key         string
	safeDefault bool
}

func (f BooleanFlag) definition() booleanDefinition {
	switch f {
	case ReleaseWebV3:
		return booleanDefinition{
			key:         "release-webv3",
			safeDefault: false,
		}
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
