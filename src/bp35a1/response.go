package bp35a1

type rp int

const (
	OK rp = iota
	FAIL
	RESULT
)

/* Response */
type Response interface {
	Type() rp
}

type response struct {
	t rp
}

func (r *response) Type() rp {
	return r.t
}

type Fail interface {
	Response
	Code() string
}

type response_fail struct {
	*response
	code string
}

func (r *response_fail) Code() string {
	return r.code
}

type Result interface {
	Response
	Result() string
}

type response_result struct {
	*response
	result string
}

func (r *response_result) Result() string {
	return r.result
}
