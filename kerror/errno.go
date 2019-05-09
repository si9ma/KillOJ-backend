// error package for killoj
package kerror

type ResponseErrno int

const (
	// 400xx : bad request
	BadRequestGeneral = ResponseErrno(40000)
	ArgValidateFail   = ResponseErrno(40001)

	// 500xx: Internal Server Error
	InternalServerErrorGeneral = ResponseErrno(50000)
)
