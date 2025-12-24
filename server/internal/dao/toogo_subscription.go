// =================================================================================
// Code generated and target for Toogo.Ai system.
// =================================================================================

package dao

import (
	"hotgo/internal/dao/internal"
)

// internalToogoSubscriptionDao is internal type for wrapping internal DAO implements.
type internalToogoSubscriptionDao = *internal.ToogoSubscriptionDao

// toogoSubscriptionDao is the data access object for table hg_toogo_subscription.
var toogoSubscriptionDao = &toogoSubscriptionDaoImpl{
	internal.NewToogoSubscriptionDao(),
}

// ToogoSubscription is the manager for table hg_toogo_subscription.
var ToogoSubscription = toogoSubscriptionDao

type toogoSubscriptionDaoImpl struct {
	internalToogoSubscriptionDao
}

