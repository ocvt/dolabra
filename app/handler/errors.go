package handler


type errInvalidPayload struct {
  s string
}

func (e *errInvalidPayload) Error() string {
  return e.s
}
