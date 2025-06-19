package ag_error

import "fmt"

type BizStatusErrorIface interface {
	BizCode() int32
	BizMessage() string
	BizExtra() map[string]string
	Error() string
}

type BizStatusError struct {
	code  int32
	msg   string
	extra map[string]string
}

// NewBizStatusError returns BizStatusErrorIface
func NewBizStatusError(code int32, msg string, extra ...map[string]string) BizStatusErrorIface {
	if len(extra) > 0 {
		return &BizStatusError{code: code, msg: msg, extra: extra[0]}
	} else {
		return &BizStatusError{code: code, msg: msg}
	}
}

/* === BizStatusError 实现 BizStatusErrorIface === */
func (e *BizStatusError) BizCode() int32 {
	return e.code
}

func (e *BizStatusError) BizMessage() string {
	return e.msg
}

func (e *BizStatusError) BizExtra() map[string]string {
	return e.extra
}
func (e *BizStatusError) Error() string {
	return fmt.Sprintf("biz error: code=%d, msg=%s", e.code, e.msg)
}

/* === BizStatusError 自定义实现 === */
