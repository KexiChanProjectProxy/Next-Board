package service

import (
	"testing"
	"time"

	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/models"

	"go.uber.org/zap"
)

// Mock repositories for testing
type mockUserRepo struct{}
type mockNodeRepo struct{}
type mockPlanRepo struct{}
type mockUsageRepo struct{}
type mockUUIDRepo struct{}

func (m *mockUserRepo) FindByID(id uint64) (*models.User, error) {
	planID := uint64(1)
	return &models.User{
		ID:     id,
		PlanID: &planID,
	}, nil
}

func (m *mockNodeRepo) FindByIDWithLabels(id uint64) (*models.Node, error) {
	return &models.Node{
		ID:             id,
		NodeMultiplier: 1.5,
		Labels: []models.Label{
			{ID: 1, Name: "Premium"},
			{ID: 2, Name: "US"},
		},
	}, nil
}

func (m *mockPlanRepo) FindByIDWithLabels(id uint64) (*models.Plan, error) {
	return &models.Plan{
		ID:             id,
		BaseMultiplier: 1.0,
		Labels: []models.Label{
			{ID: 1, Name: "Premium"},
		},
	}, nil
}

func (m *mockPlanRepo) GetAllLabelMultipliers(planID uint64) (map[uint64]float64, error) {
	return map[uint64]float64{
		1: 2.0, // Premium label has 2x multiplier
	}, nil
}

func (m *mockUsageRepo) GetCurrentPeriod(userID uint64) (*models.UsagePeriod, error) {
	return &models.UsagePeriod{
		ID:                1,
		UserID:            userID,
		PlanID:            1,
		RealBytesUp:       0,
		RealBytesDown:     0,
		BillableBytesUp:   0,
		BillableBytesDown: 0,
		IsCurrent:         true,
	}, nil
}

func (m *mockUsageRepo) IncrementUsage(userID, nodeID uint64, realUp, realDown, billableUp, billableDown uint64) error {
	return nil
}

func (m *mockUUIDRepo) GetAllUserUUIDs() (map[uint64]string, error) {
	return map[uint64]string{}, nil
}

// Test multiplier calculation
func TestCalculateMultiplier(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := &accountingService{
		userRepo:  &mockUserRepo{},
		nodeRepo:  &mockNodeRepo{},
		planRepo:  &mockPlanRepo{},
		usageRepo: &mockUsageRepo{},
		uuidRepo:  &mockUUIDRepo{},
		logger:    logger,
	}

	tests := []struct {
		name             string
		userID           uint64
		nodeID           uint64
		expectedMultiplier float64
	}{
		{
			name:   "Basic multiplier calculation",
			userID: 1,
			nodeID: 1,
			// node_multiplier (1.5) × plan_base (1.0) × label_premium (2.0) = 3.0
			expectedMultiplier: 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			multiplier, err := service.CalculateMultiplier(tt.userID, tt.nodeID)
			if err != nil {
				t.Fatalf("CalculateMultiplier() error = %v", err)
			}
			if multiplier != tt.expectedMultiplier {
				t.Errorf("CalculateMultiplier() = %v, want %v", multiplier, tt.expectedMultiplier)
			}
		})
	}
}

// Test period bounds calculation
func TestCalculatePeriodBounds(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := &accountingService{
		logger: logger,
	}

	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		resetPeriod  string
		expectedStart time.Time
		expectedEnd   time.Time
	}{
		{
			name:         "Daily period",
			resetPeriod:  "daily",
			expectedStart: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC),
		},
		{
			name:         "Weekly period",
			resetPeriod:  "weekly",
			expectedStart: time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC), // Sunday
			expectedEnd:   time.Date(2025, 1, 18, 0, 0, 0, 0, time.UTC),
		},
		{
			name:         "Monthly period",
			resetPeriod:  "monthly",
			expectedStart: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:         "Yearly period",
			resetPeriod:  "yearly",
			expectedStart: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := service.calculatePeriodBounds(now, tt.resetPeriod)
			if !start.Equal(tt.expectedStart) {
				t.Errorf("Period start = %v, want %v", start, tt.expectedStart)
			}
			if !end.Equal(tt.expectedEnd) {
				t.Errorf("Period end = %v, want %v", end, tt.expectedEnd)
			}
		})
	}
}

// Test traffic calculation with multipliers
func TestTrafficCalculation(t *testing.T) {
	tests := []struct {
		name             string
		realBytes        uint64
		nodeMultiplier   float64
		planMultiplier   float64
		labelMultiplier  float64
		expectedBillable uint64
	}{
		{
			name:             "No multipliers",
			realBytes:        1000000,
			nodeMultiplier:   1.0,
			planMultiplier:   1.0,
			labelMultiplier:  1.0,
			expectedBillable: 1000000,
		},
		{
			name:             "Node multiplier only",
			realBytes:        1000000,
			nodeMultiplier:   1.5,
			planMultiplier:   1.0,
			labelMultiplier:  1.0,
			expectedBillable: 1500000,
		},
		{
			name:             "All multipliers",
			realBytes:        1000000,
			nodeMultiplier:   1.5,
			planMultiplier:   1.2,
			labelMultiplier:  2.0,
			expectedBillable: 3600000, // 1000000 × 1.5 × 1.2 × 2.0
		},
		{
			name:             "Fractional result rounds down",
			realBytes:        1000,
			nodeMultiplier:   1.5,
			planMultiplier:   1.0,
			labelMultiplier:  1.0,
			expectedBillable: 1500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalMultiplier := tt.nodeMultiplier * tt.planMultiplier * tt.labelMultiplier
			billable := uint64(float64(tt.realBytes) * totalMultiplier)

			if billable != tt.expectedBillable {
				t.Errorf("Billable = %v, want %v", billable, tt.expectedBillable)
			}
		})
	}
}

// Benchmark multiplier calculation
func BenchmarkCalculateMultiplier(b *testing.B) {
	logger, _ := zap.NewDevelopment()

	service := &accountingService{
		userRepo:  &mockUserRepo{},
		nodeRepo:  &mockNodeRepo{},
		planRepo:  &mockPlanRepo{},
		usageRepo: &mockUsageRepo{},
		uuidRepo:  &mockUUIDRepo{},
		logger:    logger,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CalculateMultiplier(1, 1)
	}
}
