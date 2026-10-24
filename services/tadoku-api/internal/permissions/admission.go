package permissions

// Admission is the ban gate's answer for one request.
type Admission interface {
	admission()
}

type Admitted struct{}
type Banned struct{}

type AdmittedBanUnknown struct {
	Err error
}

type Unavailable struct {
	Err error
}

func (Admitted) admission()           {}
func (Banned) admission()             {}
func (AdmittedBanUnknown) admission() {}
func (Unavailable) admission()        {}
