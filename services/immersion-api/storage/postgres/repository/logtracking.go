package repository

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func trackingUnitID(tracking domain.LogTracking) uuid.NullUUID {
	if tracking.Kind != domain.LogTrackingAmountUnit && tracking.Kind != domain.LogTrackingBoth {
		return uuid.NullUUID{}
	}
	return postgres.NewNullUUID(tracking.UnitID)
}

func trackingAmount(tracking domain.LogTracking) sql.NullFloat64 {
	if tracking.Kind != domain.LogTrackingAmountUnit && tracking.Kind != domain.LogTrackingBoth {
		return sql.NullFloat64{}
	}
	return postgres.NewNullFloat64FromFloat32(tracking.Amount)
}

func trackingModifier(tracking domain.LogTracking) sql.NullFloat64 {
	if tracking.Kind != domain.LogTrackingAmountUnit && tracking.Kind != domain.LogTrackingBoth {
		return sql.NullFloat64{}
	}
	return postgres.NewNullFloat64FromFloat32(tracking.Modifier)
}

func trackingDurationSeconds(tracking domain.LogTracking) sql.NullInt32 {
	if tracking.Kind != domain.LogTrackingDuration && tracking.Kind != domain.LogTrackingBoth {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Valid: true, Int32: tracking.DurationSeconds}
}
