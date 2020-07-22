package greq

type responseBytes struct {
	Bind bool
	Bytes *[]byte
}
type responseJSON struct {
	Bind bool
	Value interface{}
}
func BindBytes(ptr *[]byte) responseBytes {
	return responseBytes{
		Bind:  true,
		Bytes: ptr,
	}
}
func BindJSON(ptr interface{}) responseJSON {
	return responseJSON{
		Bind:  true,
		Value: ptr,
	}
}
type Response struct {
	Bytes responseBytes
	JSON responseJSON
}
