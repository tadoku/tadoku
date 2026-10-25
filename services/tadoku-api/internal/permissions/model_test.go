package permissions

import "testing"

func TestModelIsValid(t *testing.T) {
	t.Parallel()

	if _, err := NewModel(sites, groups, contests, logs); err != nil {
		t.Errorf("test model: %v", err)
	}
}

func TestModelValidation(t *testing.T) {
	t.Parallel()

	owners := Relations{"owners": {Users}}
	cases := []struct {
		name       string
		namespaces []Namespace
	}{
		{"unknown relation", []Namespace{Define(Contest, nil, Permits{"edit": Related("owners")})}},
		{"unknown permit", []Namespace{Define(Contest, nil, Permits{"edit": Allowed("moderate")})}},
		{"relation and permit share a name", []Namespace{
			Define(Contest, Relations{"edit": {Users}}, Permits{"edit": Related("edit")}),
		}},
		{"via over users", []Namespace{Define(Contest, owners, Permits{"edit": Via("owners", "edit")})}},
		{"related over objects", []Namespace{
			Define(Contest, nil, nil),
			Define(Log, Relations{"contests": {ObjectOf(Contest)}}, Permits{"hide": Related("contests")}),
		}},
		{"via to a missing permit", []Namespace{
			Define(Contest, nil, nil),
			Define(Log, Relations{"contests": {ObjectOf(Contest)}}, Permits{"hide": Via("contests", "moderate")}),
		}},
		{"permit on an unknown object type", []Namespace{
			Define(Contest, nil, Permits{"edit": PermitOn(Site, siteRoot, "admin")}),
		}},
		{"permit on a missing permit", []Namespace{
			Define(Site, nil, nil),
			Define(Contest, nil, Permits{"edit": PermitOn(Site, siteRoot, "admin")}),
		}},
		{"permit on an object without an ID", []Namespace{
			sites,
			Define(Contest, nil, Permits{"edit": PermitOn(Site, "", "admin")}),
		}},
		{"single-rule combinator", []Namespace{Define(Contest, owners, Permits{"edit": AnyOf(Related("owners"))})}},
		{"unknown subject type", []Namespace{
			Define(Contest, Relations{"moderators": {SetOf(Group, "members")}}, nil),
		}},
		{"object type reserved for users", []Namespace{Define("user", Relations{"friends": {Users}}, nil)}},
		{"object type without a name", []Namespace{Define("", owners, nil)}},
		{"object type defined twice", []Namespace{Define(Contest, owners, nil), Define(Contest, nil, nil)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewModel(tc.namespaces...); err == nil {
				t.Error("invalid model passed validation")
			}
		})
	}
}
