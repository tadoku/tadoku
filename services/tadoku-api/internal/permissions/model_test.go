package permissions

import "testing"

func TestModelIsValid(t *testing.T) {
	t.Parallel()

	if _, err := NewModel(sites, groups, contests, logs); err != nil {
		t.Errorf("test model: %v", err)
	}
}

func TestDefineRejectsWrongLocalNames(t *testing.T) {
	t.Parallel()

	owners := Relations{"owners": {Users}}
	links := Relations{"contests": {ObjectOf(Contest)}}
	cases := []struct {
		name       string
		objectType ObjectType
		relations  Relations
		permits    Permits
	}{
		{"unknown relation", Contest, nil, Permits{"edit": Related("owners")}},
		{"unknown permit", Contest, nil, Permits{"edit": Allowed("moderate")}},
		{"relation and permit share a name", Contest, Relations{"edit": {Users}}, Permits{"edit": Related("edit")}},
		{"via over users", Contest, owners, Permits{"edit": Via("owners", "edit")}},
		{"via over an unknown relation", Log, nil, Permits{"hide": Via("contests", "moderate")}},
		{"related over objects", Log, links, Permits{"hide": Related("contests")}},
		{"single-rule combinator", Contest, owners, Permits{"edit": AnyOf(Related("owners"))}},
		{"missing rule", Contest, owners, Permits{"edit": AllOf(Related("owners"), nil)}},
		{"permit on an object without an ID", Contest, nil, Permits{"edit": PermitOn(Site, "", "admin")}},
		{"own subject set without a relation", Group, Relations{"members": {SetOf(Group, "admins")}}, nil},
		{"object type reserved for users", "user", Relations{"friends": {Users}}, nil},
		{"object type without a name", "", owners, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("MustDefine accepted an invalid namespace")
				}
			}()

			MustDefine(tc.objectType, tc.relations, tc.permits)
		})
	}
}

func TestModelRejectsWrongReferences(t *testing.T) {
	t.Parallel()

	contestOnly := MustDefine(Contest, nil, nil)
	cases := []struct {
		name       string
		namespaces []Namespace
	}{
		{"via to a missing permit", []Namespace{
			contestOnly,
			MustDefine(Log, Relations{"contests": {ObjectOf(Contest)}}, Permits{"hide": Via("contests", "moderate")}),
		}},
		{"permit on an unknown object type", []Namespace{
			MustDefine(Contest, nil, Permits{"edit": PermitOn(Site, siteRoot, "admin")}),
		}},
		{"permit on a missing permit", []Namespace{
			MustDefine(Site, nil, nil),
			MustDefine(Contest, nil, Permits{"edit": PermitOn(Site, siteRoot, "admin")}),
		}},
		{"unknown subject type", []Namespace{
			MustDefine(Contest, Relations{"moderators": {SetOf(Group, "members")}}, nil),
		}},
		{"unknown subject set", []Namespace{
			MustDefine(Group, nil, nil),
			MustDefine(Contest, Relations{"moderators": {SetOf(Group, "members")}}, nil),
		}},
		{"unknown linked object type", []Namespace{MustDefine(Log, Relations{"contests": {ObjectOf(Contest)}}, nil)}},
		{"object type defined twice", []Namespace{contestOnly, contestOnly}},
		{"namespace not built by MustDefine", []Namespace{{}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewModel(tc.namespaces...); err == nil {
				t.Error("invalid model passed validation")
			}
		})
	}
}
