package usecase

type HealthChecker interface {
	HealthCheck() string
}

type HealthCheck struct {
	checker HealthChecker
}

func NewHealthCheck(checker HealthChecker) *HealthCheck {
	return &HealthCheck{
		checker: checker,
	}
}

func (h *HealthCheck) HealthCheck() string {
	return h.checker.HealthCheck()
}
