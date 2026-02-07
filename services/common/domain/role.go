package domain

type Role string

var RoleAdmin = Role("admin")
var RoleUser = Role("user")
var RoleGuest = Role("guest")
var RoleBanned = Role("banned")
