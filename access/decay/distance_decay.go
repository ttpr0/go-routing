package decay

type IDistanceDecay interface {
	GetDistanceWeight(distance float32) float32

	/**
	 * Returns the maximum distance threshold.
	 * Distances higher than this get weight 0.
	 *
	 * @return maximum distance
	 */
	GetMaxDistance() float32

	/**
	 * Returns the internally used distances (currently only hybrid decay)
	 *
	 * @return distances (if available) or null
	 */
	GetDistances() []float32
}
