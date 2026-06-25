package risorchart

import "math"

// calculateStats computes mean and standard deviation
func calculateStats(values []float64) (mean, stddev float64) {
	n := float64(len(values))
	if n == 0 {
		return 0, 0
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / n

	// Calculate standard deviation
	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	stddev = math.Sqrt(variance / n)

	return mean, stddev
}

// normalPDF returns the probability density function value for a normal distribution
func normalPDF(x, mean, stddev float64) float64 {
	if stddev == 0 {
		return 0
	}
	exponent := -0.5 * math.Pow((x-mean)/stddev, 2)
	return (1.0 / (stddev * math.Sqrt(2*math.Pi))) * math.Exp(exponent)
}
