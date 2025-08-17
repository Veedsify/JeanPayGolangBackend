package services

func GetAdminDashboardStatistics() (map[string]interface{}, error) {

	stats := map[string]any{
		"totalUsers":        1000,
		"totalTransactions": 5000,
		"totalRevenue":      25000.00,
	}

	return stats, nil
}
