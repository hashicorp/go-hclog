package hclog

const inlineArgs = 10

type Record struct {
	Level    Level
	Msg      string
	CallerPc uintptr

	inline [inlineArgs]interface{}
	Args   []interface{}
}

func (r *Record) SetArgs(args []interface{}) {
	if len(args) > inlineArgs {
		r.Args = make([]interface{}, len(args))
		copy(r.Args, args)
	} else {
		copy(r.inline[:], args)
		r.Args = r.inline[:]
	}
}
