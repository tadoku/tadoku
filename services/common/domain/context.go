package domain

type TadokuContextKey string

var CtxSessionKey = TadokuContextKey("session")
var CtxIdentityKey = TadokuContextKey("identity")
