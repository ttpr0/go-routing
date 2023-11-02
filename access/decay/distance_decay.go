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
}
