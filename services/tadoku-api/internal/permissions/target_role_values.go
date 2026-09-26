package permissions

// TargetRoles are Keto facts about another user, never authorization for the actor.
type TargetRoles struct {
	Admin  bool
	Banned bool
}
