package stream

func Of[T any](s []T) Stream[T] {
	return newPipelineStream(s)
}
