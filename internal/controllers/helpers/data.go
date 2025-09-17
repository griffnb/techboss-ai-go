package helpers

type DataWrap[T any] struct {
	Data T `json:"data"`
}
